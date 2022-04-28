package backpressure

import (
	"runtime/debug"
	"sync/atomic"

	"gitlab.stageoffice.ru/UCS-COMMON/gaben"
	servus_stats "gitlab.stageoffice.ru/UCS-COMMON/schemagen-go/v41/servus/stats/v1"

	"github.com/pkg/errors"
)

var _ Operator = (*operatorImpl)(nil)

type operatorImpl struct {
	*throttler
	lastControlParameters atomic.Value
	logger                gaben.Logger
}

func (b *operatorImpl) GetStats() *servus_stats.GoMemLimiterStats_BackpressureStats {
	result := &servus_stats.GoMemLimiterStats_BackpressureStats{
		Throttling: b.throttler.getStats(),
	}

	lastControlParameters := b.lastControlParameters.Load()
	if lastControlParameters != nil {
		result.ControlParameters = lastControlParameters.(*ControlParameters).ToProtobuf()
	}

	return result
}

func (b *operatorImpl) SetControlParameters(value *ControlParameters) error {
	old := b.lastControlParameters.Swap(value)
	if old != nil {
		// если управляющие параметры не изменились, ничего не делаем
		if value.equalsTo(old.(*ControlParameters)) {
			return nil
		}
	}

	// регулируем количество поступающих запросов
	if err := b.throttler.setThreshold(value.ThrottlingPercentage); err != nil {
		return errors.Wrap(err, "throttler set threshold")
	}

	// и интенсивность сбора мусора
	debug.SetGCPercent(value.GOGC)

	b.logger.Info("control parameters changed", value.ToGaben()...)

	return nil
}

// NewOperator - конструктор нового оператора.
func NewOperator(logger gaben.Logger) Operator {
	return &operatorImpl{
		logger:    logger,
		throttler: newThrottler(),
	}
}
