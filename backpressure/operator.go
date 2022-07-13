/*
 * Copyright (c) New Cloud Technologies, Ltd. 2013-2022.
 * Author: Vitaly Isaev <vitaly.isaev@myoffice.team>
 * License: https://github.com/newcloudtechnologies/memlimiter/blob/master/LICENSE
 */

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
	notificationChan      chan<- *stats.MemLimiterStats
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
		// if control parameters didn't change, we do nothing
		if value.EqualsTo(old.(*stats.ControlParameters)) {
			return nil
		}
	}

	// set the share of the requests that have to be throttled
	if err := b.throttler.setThreshold(value.ThrottlingPercentage); err != nil {
		return errors.Wrap(err, "throttler set threshold")
	}

	// tune GC pace
	debug.SetGCPercent(value.GOGC)

	b.logger.Info("control parameters changed", value.ToKeysAndValues()...)

	// notify client about statistics change
	if b.notificationChan != nil {
		backpressureStats, err := b.GetStats()
		if err != nil {
			return errors.Wrap(err, "get stats")
		}

		memLimiterStats := &stats.MemLimiterStats{
			Controller:   value.ControllerStats,
			Backpressure: backpressureStats,
		}

		// if client's not ready to read, omit it
		select {
		case b.notificationChan <- memLimiterStats:
		default:
		}
	}

	return nil
}

// NewOperator builds new Operator.
func NewOperator(logger logr.Logger, options ...Option) Operator {
	out := &operatorImpl{
		logger:    logger,
		throttler: newThrottler(),
	}

	//nolint:gocritic
	for _, op := range options {
		switch t := op.(type) {
		case *notificationsOption:
			out.notificationChan = t.val
		}
	}

	return out
}
