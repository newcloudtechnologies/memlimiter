/*
 * Copyright (c) New Cloud Technologies, Ltd. 2013-2022.
 * Author: Vitaly Isaev <vitaly.isaev@myoffice.team>
 * License: https://github.com/newcloudtechnologies/memlimiter/blob/master/LICENSE
 */

package nextgc

import (
	"fmt"
	"math"
	"time"

	"github.com/go-logr/logr"
	"github.com/newcloudtechnologies/memlimiter/stats"
	"github.com/newcloudtechnologies/memlimiter/utils/breaker"

	"github.com/newcloudtechnologies/memlimiter/backpressure"
	"github.com/newcloudtechnologies/memlimiter/controller"
	memlimiter_utils "github.com/newcloudtechnologies/memlimiter/utils"
)

// controllerImpl - in some early versions this class was designed to be used as a classical PID-controller
// described in control theory. But currently it has only proportional (P) component, and the proportionality
// is non-linear (see component_p.go). It looks like integral (I) component will never be implemented.
// But the differential controller (D) still may be implemented in future if we face self-oscillation.
type controllerImpl struct {
	input  stats.ServiceStatsSubscription // input: service tracker subscription.
	output backpressure.Operator          // output: write control parameters here

	// Controller components:
	// 1. proportional component.
	componentP *componentP

	// cached values, describing the actual state of the controller:
	pValue            float64                  // proportional component's output
	sumValue          float64                  // final output
	goAllocLimit      uint64                   // memory budget [bytes]
	utilization       float64                  // memory budget utilization [percents]
	rss               uint64                   // physical memory actual consumption
	consumptionReport *stats.ConsumptionReport // latest special memory consumers report
	controlParameters *stats.ControlParameters // latest control parameters value

	getStatsChan chan *getStatsRequest

	cfg     *ControllerConfig
	logger  logr.Logger
	breaker *breaker.Breaker
}

// getStatsRequest is a request to get the controller stats.
type getStatsRequest struct {
	result chan *stats.ControllerStats
}

// NewControllerFromConfig builds new controller.
func NewControllerFromConfig(
	logger logr.Logger,
	cfg *ControllerConfig,
	serviceStatsSubscription stats.ServiceStatsSubscription,
	backpressureOperator backpressure.Operator,
) (controller.Controller, error) {
	c := &controllerImpl{
		input:      serviceStatsSubscription,
		output:     backpressureOperator,
		componentP: newComponentP(logger, cfg.ComponentProportional),
		pValue:     0,
		sumValue:   0,
		controlParameters: &stats.ControlParameters{
			GOGC:                 backpressure.DefaultGOGC,
			ThrottlingPercentage: backpressure.NoThrottling,
		},
		getStatsChan: make(chan *getStatsRequest),
		cfg:          cfg,
		logger:       logger,
		breaker:      breaker.NewBreakerWithInitValue(1),
	}

	// initialize backpressure operator with default control signal
	err := c.applyControlValue()
	if err != nil {
		return nil, fmt.Errorf("apply control value: %w", err)
	}

	go c.loop()

	return c, nil
}

// GetStats returns the current controller stats.
func (c *controllerImpl) GetStats() (*stats.ControllerStats, error) {
	req := &getStatsRequest{result: make(chan *stats.ControllerStats, 1)}

	select {
	case c.getStatsChan <- req:
	case <-c.breaker.Done():
		return nil, fmt.Errorf("breaker err: %w", c.breaker.Err())
	}

	select {
	case resp := <-req.result:
		return resp, nil
	case <-c.breaker.Done():
		return nil, fmt.Errorf("breaker err: %w", c.breaker.Err())
	}
}

// Quit gracefully stops the controller.
func (c *controllerImpl) Quit() {
	c.breaker.ShutdownAndWait()
}

// respondWith responds with the controller stats.
func (r *getStatsRequest) respondWith(resp *stats.ControllerStats) {
	r.result <- resp
}

// loop is the main loop of the controller.
func (c *controllerImpl) loop() {
	defer c.breaker.Dec()

	ticker := time.NewTicker(c.cfg.Period.Duration)
	defer ticker.Stop()

	for {
		select {
		case serviceStats := <-c.input.Updates():
			// Update controller state every time we receive the actual tracker about the process.
			err := c.updateState(serviceStats)
			if err != nil {
				c.logger.Error(err, "update state")
			}
		case <-ticker.C:
			// Generate control parameters based on the most recent state and send it to the backpressure operator.
			err := c.applyControlValue()
			if err != nil {
				c.logger.Error(err, "apply control value")
			}
		case req := <-c.getStatsChan:
			req.respondWith(c.aggregateStats())
		case <-c.breaker.Done():
			return
		}
	}
}

// updateState updates the controller state.
func (c *controllerImpl) updateState(serviceStats stats.ServiceStats) error {
	// Extract the latest report on special memory consumers if there are any.
	c.consumptionReport = serviceStats.ConsumptionReport()

	c.updateUtilization(serviceStats)

	err := c.updateControlValues()
	if err != nil {
		return fmt.Errorf("update control values: %w", err)
	}

	c.updateControlParameters()

	return nil
}

// updateUtilization updates the controller utilization.
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

	goAllocLimit, budgetOK := c.computeGoAllocLimit(cgoAllocs)
	c.goAllocLimit = goAllocLimit

	// Memory utilization is defined as the relation of NextGC value to the Go allocation limit.
	// If NextGC becomes higher than the allocation limit, the GC will never run, because
	// OOM will happen first. That's why we need to push away Go process from the allocation limit.
	if !budgetOK {
		// If non-Go allocations already exhausted the RSS budget, force controller to
		// apply conservative parameters without producing infinities/NaN in stats output.
		c.utilization = exhaustedBudgetUtilization
		c.rss = serviceStats.RSS()

		return
	}

	c.utilization = float64(serviceStats.NextGC()) / float64(c.goAllocLimit)

	// Just for the history, save actual RSS value
	c.rss = serviceStats.RSS()
}

// computeGoAllocLimit computes Go allocations budget from total RSS limit and cgo consumption.
// bool result indicates whether the resulting budget is valid and non-exhausted.
func (c *controllerImpl) computeGoAllocLimit(cgoAllocs uint64) (uint64, bool) {
	rssLimit := c.cfg.RSSLimit.Value

	if rssLimit == 0 {
		return 0, false
	}

	if cgoAllocs >= rssLimit {
		return 1, false
	}

	return rssLimit - cgoAllocs, true
}

// updateControlValues updates the controller control values.
func (c *controllerImpl) updateControlValues() error {
	var err error

	c.pValue, err = c.componentP.value(c.utilization)
	if err != nil {
		return fmt.Errorf("component proportional value: %w", err)
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

// updateControlParameters updates the controller control parameters.
func (c *controllerImpl) updateControlParameters() {
	c.controlParameters = &stats.ControlParameters{}
	c.updateControlParameterGOGC()
	c.updateControlParameterThrottling()

	c.controlParameters.ControllerStats = c.aggregateStats()
}

const (
	// percents is a constant for converting float64 to uint32.
	percents = 100
	// exhaustedBudgetUtilization is a finite marker value greater than 1 used
	// when cgo allocations fully consume RSS budget.
	// This avoids Inf/NaN in stats output and guarantees "red zone" behavior.
	exhaustedBudgetUtilization = 1.01
)

// updateControlParameterGOGC updates the controller control parameter GOGC.
func (c *controllerImpl) updateControlParameterGOGC() {
	// Control parameters are set to defaults in the "green zone".
	if uint32(c.utilization*percents) < c.cfg.DangerZoneGOGC {
		c.controlParameters.GOGC = backpressure.DefaultGOGC

		return
	}

	// Control parameters are more conservative in the "red zone".
	roundedValue := uint32(math.Round(c.sumValue))
	gogc := int(backpressure.DefaultGOGC - roundedValue)

	minGOGC := c.cfg.MinGOGC
	if minGOGC == 0 {
		minGOGC = defaultMinGOGC
	}

	if gogc < minGOGC {
		gogc = minGOGC
	}

	c.controlParameters.GOGC = gogc
}

// updateControlParameterThrottling updates the controller control parameter throttling.
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

// applyControlValue applies the controller control value.
func (c *controllerImpl) applyControlValue() error {
	err := c.output.SetControlParameters(c.controlParameters)
	if err != nil {
		return fmt.Errorf("set control parameters: %v: %w", c.controlParameters, err)
	}

	return nil
}

// aggregateStats aggregates the controller stats.
func (c *controllerImpl) aggregateStats() *stats.ControllerStats {
	res := &stats.ControllerStats{
		MemoryBudget: &stats.MemoryBudgetStats{
			RSSActual:    c.rss,
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
