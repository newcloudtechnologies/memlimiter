/*
 * Copyright (c) New Cloud Technologies, Ltd. 2013-2022.
 * Author: Vitaly Isaev <vitaly.isaev@myoffice.team>
 * License: https://github.com/newcloudtechnologies/memlimiter/blob/master/LICENSE
 */

package nextgc

import (
	"fmt"
	"math"
	"sync"

	"github.com/go-logr/logr"
)

// The proportional component of the controller.
type componentP struct {
	valueSmoother *emaSmoother
	cfg           *ComponentProportionalConfig
	logger        logr.Logger
}

func (c *componentP) value(utilization float64) (float64, error) {
	if c.valueSmoother != nil {
		valueEMA, err := c.valueEMA(utilization)
		if err != nil {
			return math.NaN(), fmt.Errorf("value EMA: %w", err)
		}

		return valueEMA, nil
	}

	valueRaw, err := c.valueRaw(utilization)
	if err != nil {
		return math.NaN(), fmt.Errorf("value raw: %w", err)
	}

	return valueRaw, nil
}

func (c *componentP) valueRaw(utilization float64) (float64, error) {
	if utilization < 0 {
		return math.NaN(), fmt.Errorf("value is undefined if memory usage = %v", utilization)
	}

	if utilization >= 1 {
		// In theory, values >= 1 are impossible, but in practice sometimes we face small exceeding of the upper bound (< 1.1).
		// This needs to be investigated later.
		const maxReasonableOutput = 100

		c.logger.Info(
			"not a good value for memory usage, cutting output value",
			"memory_usage", utilization,
			"output_value", maxReasonableOutput,
		)

		return maxReasonableOutput, nil
	}

	// The closer the memory usage value is to 100%, the stronger the control signal.
	return c.cfg.Coefficient * (1 / (1 - utilization)), nil
}

func (c *componentP) valueEMA(utilization float64) (float64, error) {
	valueRaw, err := c.valueRaw(utilization)
	if err != nil {
		return 0, fmt.Errorf("value raw: %w", err)
	}

	return c.valueSmoother.Update(valueRaw), nil
}

func newComponentP(logger logr.Logger, cfg *ComponentProportionalConfig) *componentP {
	out := &componentP{
		logger: logger,
		cfg:    cfg,
	}

	if cfg.WindowSize != 0 {
		// alpha is a smoothing coefficient describing the degree of weighting decrease;
		// the lesser the alpha is, the higher the impact of the elder historical values on the resulting value.
		// alpha is choosed empirically, but can depend on a window size, like here:
		// https://en.wikipedia.org/wiki/Moving_average#Relationship_between_SMA_and_EMA
		//nolint:gomnd
		alpha := 2 / (float64(cfg.WindowSize + 1))

		out.valueSmoother = newEMASmoother(alpha)
	}

	return out
}

type emaSmoother struct {
	mu          sync.Mutex
	alpha       float64
	initialized bool
	value       float64
}

func newEMASmoother(alpha float64) *emaSmoother {
	return &emaSmoother{alpha: alpha}
}

func (e *emaSmoother) Update(v float64) float64 {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.initialized {
		e.value = v
		e.initialized = true

		return e.value
	}

	e.value = e.alpha*v + (1-e.alpha)*e.value

	return e.value
}
