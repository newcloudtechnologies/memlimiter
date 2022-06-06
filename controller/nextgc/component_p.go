package nextgc

import (
	"math"

	"github.com/go-logr/logr"
	metrics "github.com/rcrowley/go-metrics"

	"github.com/pkg/errors"
)

// The proportional component of the controller.
type componentP struct {
	logger     logr.Logger
	lastValues metrics.Sample
	cfg        *ComponentProportionalConfig
}

func (c *componentP) value(memoryUsage float64) (float64, error) {
	if c.lastValues != nil {
		valueEMA, err := c.valueEMA(memoryUsage)
		if err != nil {
			return math.NaN(), errors.Wrap(err, "value EMA")
		}

		return valueEMA, nil
	}

	valueRaw, err := c.valueRaw(memoryUsage)
	if err != nil {
		return math.NaN(), errors.Wrap(err, "value raw")
	}

	return valueRaw, nil
}

func (c *componentP) valueRaw(memoryUsage float64) (float64, error) {
	if memoryUsage < 0 {
		return math.NaN(), errors.Errorf("value is undefined if memory usage = %v", memoryUsage)
	}

	if memoryUsage >= 1 {
		// In theory, values >= 1 cannot be reached, but in practice sometimes we face
		// with sim
		// NOTE:
		// Теоретически значения >= 1 недостижимы, но на практике встречаются ситуации с небольшим преувеличением лимита (< 1.1),
		// во всяком случае, встречались раньше, когда MemLimiter таргетировал не RSS utilization, a Memory Budget utilization.
		const maxReasonableOutput = 100

		c.logger.Info(
			"not a good value for memory usage, cutting output value",
			"memory_usage", memoryUsage,
			"output_value", maxReasonableOutput,
		)

		return maxReasonableOutput, nil
	}

	return c.cfg.Coefficient * (1 / (1 - memoryUsage)), nil
}

func (c *componentP) valueEMA(memoryUsage float64) (float64, error) {
	valueRaw, err := c.valueRaw(memoryUsage)
	if err != nil {
		return 0, errors.Wrap(err, "value raw")
	}

	// Эта либа работает только с интами, искать лучшую пока нет времени
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
		// alpha - сглаживающая константа, чем она меньше, тем больше влияние исторических величин
		// на итоговое значение. Выбирается эмпирически, но можно связать с окном осреднения,
		// я взял формулу отсюда:
		// https://ru.wikipedia.org/wiki/%D0%A1%D0%BA%D0%BE%D0%BB%D1%8C%D0%B7%D1%8F%D1%89%D0%B0%D1%8F_%D1%81%D1%80%D0%B5%D0%B4%D0%BD%D1%8F%D1%8F
		//nolint:gomnd
		alpha := 2 / (float64(cfg.WindowSize + 1))

		out.lastValues = metrics.NewExpDecaySample(int(cfg.WindowSize), alpha)
	}

	return out
}
