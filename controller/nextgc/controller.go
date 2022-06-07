package nextgc

import (
	"math"
	"time"

	"github.com/go-logr/logr"
	"github.com/newcloudtechnologies/memlimiter/stats"
	"github.com/newcloudtechnologies/memlimiter/utils/breaker"
	"github.com/pkg/errors"

	"github.com/newcloudtechnologies/memlimiter/backpressure"
	"github.com/newcloudtechnologies/memlimiter/controller"
	memlimiter_utils "github.com/newcloudtechnologies/memlimiter/utils"
)

// controllerImpl - in some early versions this class was designed to be used as a classical PID-controller
// described in control theory. But currently it has only proportional (P) component, and the proportionality
// is non-linear (see component_p.go). It looks like integral (I) component will never be implemented.
// But the differential controller (D) still may be implemented in future if we face self-oscillation.
//
//nolint:govet
type controllerImpl struct {
	input                 stats.ServiceStatsSubscription         // input: service stats subscription.
	backpressureOperator  backpressure.Operator                  // output: write control parameters here
	applicationTerminator memlimiter_utils.ApplicationTerminator // output: used in case of emergency stop

	// Controller components:
	// 1. proportional component.
	componentP *componentP

	// cached values, describing the actual state of the controller:
	pValue            float64                  // proportional component's output
	sumValue          float64                  // final output
	goAllocLimit      uint64                   // memory budget [bytes]
	utilization       float64                  // memory budget utilization [percents]
	consumptionReport *stats.ConsumptionReport // latest special memory consumers report
	controlParameters *stats.ControlParameters // latest control parameters value

	getStatsChan chan *getStatsRequest

	cfg     *ControllerConfig
	logger  logr.Logger
	breaker *breaker.Breaker
}

type getStatsRequest struct {
	result chan *stats.ControllerStats
}

func (r *getStatsRequest) respondWith(resp *stats.ControllerStats) {
	r.result <- resp
}

func (c *controllerImpl) GetStats() (*stats.ControllerStats, error) {
	req := &getStatsRequest{result: make(chan *stats.ControllerStats, 1)}

	select {
	case c.getStatsChan <- req:
	case <-c.breaker.Done():
		return nil, errors.Wrap(c.breaker.Err(), "breaker err")
	}

	select {
	case resp := <-req.result:
		return resp, nil
	case <-c.breaker.Done():
		return nil, errors.Wrap(c.breaker.Err(), "breaker err")
	}
}

func (c *controllerImpl) loop() {
	defer c.breaker.Dec()

	ticker := time.NewTicker(c.cfg.Period.Duration)
	defer ticker.Stop()

	for {
		select {
		case serviceStats := <-c.input.Updates():
			// Update controller state every time we receive the actual stats about the process.
			if err := c.updateState(serviceStats); err != nil {
				c.logger.Error(err, "update state")
				// Impossibility to compute control parameters is a fatal error,
				// terminate app using side-effect.
				c.applicationTerminator.Terminate(err)

				return
			}
		case <-ticker.C:
			// Generate control parameters based on the most recent state and send it to the backpressure operator.
			if err := c.applyControlValue(); err != nil {
				c.logger.Error(err, "apply control value")
				// Impossibility to apply control parameters is a fatal error,
				// terminate app using side-effect.
				c.applicationTerminator.Terminate(err)

				return
			}
		case req := <-c.getStatsChan:
			req.respondWith(c.aggregateStats())
		case <-c.breaker.Done():
			return
		}
	}
}

func (c *controllerImpl) updateState(serviceStats stats.ServiceStats) error {
	// Extract latest report on special memory consumers if any.
	var err error

	c.consumptionReport, err = serviceStats.PredefinedConsumers()
	if err != nil {
		return errors.Wrap(err, "predefined consumers")
	}

	c.updateUtilization(serviceStats)

	if err := c.updateControlValues(); err != nil {
		return errors.Wrap(err, "update control values")
	}

	c.updateControlParameters()

	return nil
}

func (c *controllerImpl) updateUtilization(serviceStats stats.ServiceStats) {
	// The process memory (roughly) consists of two main parts:
	// 1. Allocations managed by Go runtime.
	// 2. Allocations made beyond CGO border.
	//
	// We can only affect the Go allocation. CGO allocations are out of the scope.
	// To compute the amount of memory available for allocations in Go,
	// we subtract known CGO allocations from the common memory bugdet.
	// If CGO allocations grow, Go allocation have to shrink.
	var cgoAllocs uint64

	if c.consumptionReport != nil {
		for _, value := range c.consumptionReport.Cgo {
			cgoAllocs += value
		}
	}

	c.goAllocLimit = c.cfg.RSSLimit.Value - cgoAllocs

	// Memory utilization is defined as the relation of NextGC value to the Go allocation limit.
	// If NextGC becomes higher than the allocation limit, the GC will never run, because
	// OOM will happen first. That's why we need to push away Go process from the allocation limit.
	c.utilization = float64(serviceStats.NextGC()) / float64(c.goAllocLimit)
}

func (c *controllerImpl) updateControlValues() error {
	var err error

	c.pValue, err = c.componentP.value(c.utilization)
	if err != nil {
		return errors.Wrap(err, "component proportional value")
	}

	// TODO: if new components appear, summarize their outputs here:
	c.sumValue = c.pValue

	// Saturate controller output so that the control parameters are not too radical.
	// Details:
	// https://en.wikipedia.org/wiki/Saturation_arithmetic
	// https://habr.com/ru/post/345972/
	const (
		lowerBound = 0
		upperBound = 99 // this otherwise GOGC will turn to zero
	)

	c.sumValue = memlimiter_utils.ClampFloat64(c.sumValue, lowerBound, upperBound)

	return nil
}

func (c *controllerImpl) updateControlParameters() {
	c.controlParameters = &stats.ControlParameters{}
	c.updateControlParameterGOGC()
	c.updateControlParameterThrottling()
}

const percents = 100

func (c *controllerImpl) updateControlParameterGOGC() {
	// Control parameters are set to defaults in the "green zone".
	if uint32(c.utilization*percents) < c.cfg.DangerZoneGOGC {
		c.controlParameters.GOGC = backpressure.DefaultGOGC

		return
	}

	// Control parameters are more conservative in the "red zone".
	roundedValue := uint32(math.Round(c.sumValue))
	c.controlParameters.GOGC = int(backpressure.DefaultGOGC - roundedValue)
}

func (c *controllerImpl) updateControlParameterThrottling() {
	// Disable throttling in the "green zone".
	if uint32(c.utilization*percents) < c.cfg.DangerZoneThrottling {
		c.controlParameters.ThrottlingPercentage = backpressure.NoThrottling

		return
	}

	// Control parameters are more conservative in the "red zone".
	roundedValue := uint32(math.Round(c.sumValue))
	c.controlParameters.ThrottlingPercentage = roundedValue
}

func (c *controllerImpl) applyControlValue() error {
	if c.controlParameters == nil {
		c.logger.Info("control parameters are not ready yet")

		return nil
	}

	if err := c.backpressureOperator.SetControlParameters(c.controlParameters); err != nil {
		return errors.Wrapf(err, "set control parameters: %v", c.controlParameters)
	}

	return nil
}

func (c *controllerImpl) aggregateStats() *stats.ControllerStats {
	res := &stats.ControllerStats{
		MemoryBudget: &stats.MemoryBudgetStats{
			RSSLimit:     c.cfg.RSSLimit.Value,
			GoAllocLimit: c.goAllocLimit,
			Utilization:  c.utilization,
		},
		NextGC: &stats.ControllerNextGCStats{
			P:      c.pValue,
			Output: c.sumValue,
		},
	}

	if c.consumptionReport != nil {
		res.MemoryBudget.SpecialConsumers = &stats.SpecialConsumersStats{}
		res.MemoryBudget.SpecialConsumers.Go = c.consumptionReport.Go
		res.MemoryBudget.SpecialConsumers.Cgo = c.consumptionReport.Cgo
	}

	return res
}

// Quit gracefully stops the controller.
func (c *controllerImpl) Quit() {
	c.breaker.ShutdownAndWait()
}

// NewControllerFromConfig builds new controller.
func NewControllerFromConfig(
	logger logr.Logger,
	cfg *ControllerConfig,
	input stats.ServiceStatsSubscription,
	backpressureOperator backpressure.Operator,
	applicationTerminator memlimiter_utils.ApplicationTerminator,
) controller.Controller {
	c := &controllerImpl{
		input:                 input,
		backpressureOperator:  backpressureOperator,
		componentP:            newComponentP(logger, cfg.ComponentProportional),
		pValue:                0,
		sumValue:              0,
		applicationTerminator: applicationTerminator,
		controlParameters:     nil,
		getStatsChan:          make(chan *getStatsRequest),
		cfg:                   cfg,
		logger:                logger,
		breaker:               breaker.NewBreakerWithInitValue(1),
	}

	go c.loop()

	return c
}
