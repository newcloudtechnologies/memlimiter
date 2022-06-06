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
	input                 stats.Subscription                     // input: service stats subscription.
	consumptionReporter   memlimiter_utils.ConsumptionReporter   // input: information about special memory consumers.
	backpressureOperator  backpressure.Operator                  // output: write control parameters here
	applicationTerminator memlimiter_utils.ApplicationTerminator // output: used in case of emergency stop

	// Controller components:
	// 1. proportional component.
	componentP *componentP

	// cached values, describing the actual state of the controller:
	pValue            float64                             // proportional component's output
	sumValue          float64                             // final output
	goAllocLimit      uint64                              // memory budget [bytes]
	utilization       float64                             // memory budget utilization [percents]
	consumptionReport *memlimiter_utils.ConsumptionReport // latest special memory consumers report
	controlParameters *stats.ControlParameters            // latest control parameters value

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
		return nil, c.breaker.Err()
	}

	select {
	case resp := <-req.result:
		return resp, nil
	case <-c.breaker.Done():
		return nil, c.breaker.Err()
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

func (c *controllerImpl) updateState(serviceStats *stats.ServiceStats) error {
	// Extract latest report on special memory consumers if any.
	if c.consumptionReporter != nil {
		var err error

		c.consumptionReport, err = c.consumptionReporter.PredefinedConsumers(serviceStats.Custom)
		if err != nil {
			return errors.Wrap(err, "predefined consumers")
		}
	}

	c.updateUtilization(serviceStats)

	if err := c.updateControlValues(); err != nil {
		return errors.Wrap(err, "update control values")
	}

	c.updateControlParameters()

	return nil
}

func (c *controllerImpl) updateUtilization(serviceStats *stats.ServiceStats) {
	// To figure out, how much memory can be used in Go.
	// Чтобы понять, сколько памяти можно аллоцировать на нужды Go,
	// требуется вычесть из общего лимита на RSS память, в явном виде потраченную в CGO.
	// Иными словами, если аллокации в CGO будут расти, то аллокации в Go должны ужиматься.
	var cgoAllocs uint64

	if c.consumptionReport != nil {
		for _, value := range c.consumptionReport.Cgo {
			cgoAllocs += value
		}
	}

	c.goAllocLimit = c.cfg.RSSLimit.Value - cgoAllocs

	c.utilization = float64(serviceStats.NextGC) / float64(c.goAllocLimit)
}

func (c *controllerImpl) updateControlValues() error {
	var err error

	c.pValue, err = c.componentP.value(c.utilization)
	if err != nil {
		return errors.Wrap(err, "component p value")
	}

	// TODO: при появлении новых компонент суммировать их значения здесь:
	c.sumValue = c.pValue

	// Сатурация выхода регулятора, чтобы эффект вышел не слишком сильным
	// Детали:
	// https://en.wikipedia.org/wiki/Saturation_arithmetic
	// https://habr.com/ru/post/345972/
	const (
		lowerBound = 0
		upperBound = 99 // иначе GOGC превратится в 0
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
	// 	В зелёной зоне воздействие сбрасывается до дефолтов.
	if uint32(c.utilization*percents) < c.cfg.DangerZoneGOGC {
		c.controlParameters.GOGC = backpressure.DefaultGOGC

		return
	}

	// В красной зоне для GC устанавливаются более консервативные настройки
	roundedValue := uint32(math.Round(c.sumValue))
	c.controlParameters.GOGC = int(backpressure.DefaultGOGC - roundedValue)
}

func (c *controllerImpl) updateControlParameterThrottling() {
	// 	В зелёной зоне воздействие сбрасывается до дефолтов.
	if uint32(c.utilization*percents) < c.cfg.DangerZoneThrottling {
		c.controlParameters.ThrottlingPercentage = backpressure.NoThrottling

		return
	}

	// В красной зоне для GC устанавливаются более консервативные настройки
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

// Quit корректно завершает работу.
func (c *controllerImpl) Quit() {
	c.breaker.Shutdown()
	c.breaker.Wait()
}

// NewControllerFromConfig - контроллер регулятора.
func NewControllerFromConfig(
	logger logr.Logger,
	cfg *ControllerConfig,
	input stats.Subscription,
	consumptionReporter memlimiter_utils.ConsumptionReporter,
	backpressureOperator backpressure.Operator,
	applicationTerminator memlimiter_utils.ApplicationTerminator,
) controller.Controller {
	c := &controllerImpl{
		input:                 input,
		consumptionReporter:   consumptionReporter,
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
