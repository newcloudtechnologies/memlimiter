package backpressure

import (
	"runtime/debug"
	"sync/atomic"

	"github.com/go-logr/logr"
	"github.com/newcloudtechnologies/memlimiter/stats"
	"github.com/pkg/errors"
)

var _ Operator = (*operatorImpl)(nil)

type operatorImpl struct {
	*throttler
	lastControlParameters atomic.Value
	logger                logr.Logger
}

func (b *operatorImpl) GetStats() (*stats.BackpressureStats, error) {
	result := &stats.BackpressureStats{
		Throttling: b.throttler.getStats(),
	}

	lastControlParameters := b.lastControlParameters.Load()
	if lastControlParameters != nil {
		var ok bool

		result.ControlParameters, ok = lastControlParameters.(*stats.ControlParameters)
		if !ok {
			return nil, errors.Errorf("ivalid type cast (%T)", lastControlParameters)
		}
	}

	return result, nil
}

func (b *operatorImpl) SetControlParameters(value *stats.ControlParameters) error {
	old := b.lastControlParameters.Swap(value)
	if old != nil {
		// если управляющие параметры не изменились, ничего не делаем
		if value.EqualsTo(old.(*stats.ControlParameters)) {
			return nil
		}
	}

	// регулируем количество поступающих запросов
	if err := b.throttler.setThreshold(value.ThrottlingPercentage); err != nil {
		return errors.Wrap(err, "throttler set threshold")
	}

	// и интенсивность сбора мусора
	debug.SetGCPercent(value.GOGC)

	b.logger.Info("control parameters changed", value.ToKeysAndValues()...)

	return nil
}

// NewOperator - конструктор нового оператора.
func NewOperator(logger logr.Logger) Operator {
	return &operatorImpl{
		logger:    logger,
		throttler: newThrottler(),
	}
}
