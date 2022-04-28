/*
 * Copyright (c) New Cloud Technologies, Ltd. 2013-2022.
 * Author: Vitaly Isaev <vitaly.isaev@myoffice.team>
 * License: https://github.com/newcloudtechnologies/memlimiter/blob/master/LICENSE
 */

package nextgc

import (
	"math"

	"github.com/go-logr/logr"
	metrics "github.com/rcrowley/go-metrics"

	"github.com/pkg/errors"
)

// The proportional component of the controller.
type componentP struct {
	lastValues metrics.Sample
	cfg        *ComponentProportionalConfig
	logger     logr.Logger
}

func (c *componentP) value(utilization float64) (float64, error) {
	if c.lastValues != nil {
		valueEMA, err := c.valueEMA(utilization)
		if err != nil {
			return math.NaN(), errors.Wrap(err, "value EMA")
		}

		return valueEMA, nil
	}

	valueRaw, err := c.valueRaw(utilization)
	if err != nil {
		return math.NaN(), errors.Wrap(err, "value raw")
	}

	return valueRaw, nil
}

func (c *componentP) valueRaw(utilization float64) (float64, error) {
	if utilization < 0 {
		return math.NaN(), errors.Errorf("value is undefined if memory usage = %v", utilization)
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
		return 0, errors.Wrap(err, "value raw")
	}

	// TODO: need to find statistical library working with floats to make this conversion unnecessary
	const reductionFactor = 100

	c.lastValues.Update(int64(valueRaw * reductionFactor))

	return c.lastValues.Mean() / reductionFactor, nil
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

		out.lastValues = metrics.NewExpDecaySample(int(cfg.WindowSize), alpha)
	}

	return out
}
