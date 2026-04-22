/*
 * Copyright (c) New Cloud Technologies, Ltd. 2013-2022.
 * Author: Vitaly Isaev <vitaly.isaev@myoffice.team>
 * License: https://github.com/newcloudtechnologies/memlimiter/blob/master/LICENSE
 */

package nextgc

import (
	"fmt"
	"math"

	"github.com/go-logr/logr"
	"github.com/newcloudtechnologies/memlimiter/utils"
)

// The proportional component of the controller.
type componentP struct {
	// valueSmoother is a smoother for the raw proportional signal.
	valueSmoother *utils.EMASmoother
	// cfg is the configuration for the proportional component.
	cfg *ComponentProportionalConfig
	// logger is the logger for the proportional component.
	logger logr.Logger
}

// newComponentP creates a new proportional component.
func newComponentP(logger logr.Logger, cfg *ComponentProportionalConfig) *componentP {
	out := &componentP{
		logger: logger,
		cfg:    cfg,
	}

	if cfg.WindowSize != 0 {
		// We smooth the raw proportional signal because memory usage is noisy:
		// short spikes should not immediately trigger aggressive control actions.
		//
		// EMA formula:
		//   S_t = alpha*X_t + (1-alpha)*S_{t-1}
		//
		// To approximate a simple moving average window of size N, we use:
		//   alpha = 2 / (N + 1)
		//
		// Larger window -> smaller alpha -> smoother but slower reaction.
		//nolint:gomnd
		alpha := 2 / (float64(cfg.WindowSize + 1))

		out.valueSmoother = utils.NewEMASmoother(alpha)
	}

	return out
}

// value returns the proportional component's output.
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

// valueRaw returns the raw proportional component's output.
func (c *componentP) valueRaw(utilization float64) (float64, error) {
	if math.IsNaN(utilization) {
		return math.NaN(), fmt.Errorf("value is undefined if memory usage = %v", utilization)
	}

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

// valueEMA returns the exponential moving average of the raw proportional component's output.
func (c *componentP) valueEMA(utilization float64) (float64, error) {
	valueRaw, err := c.valueRaw(utilization)
	if err != nil {
		return 0, fmt.Errorf("value raw: %w", err)
	}

	return c.valueSmoother.Update(valueRaw), nil
}
