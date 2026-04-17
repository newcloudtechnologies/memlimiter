/*
 * Copyright (c) New Cloud Technologies, Ltd. 2013-2022.
 * Author: Vitaly Isaev <vitaly.isaev@myoffice.team>
 * License: https://github.com/newcloudtechnologies/memlimiter/blob/master/LICENSE
 */

package backpressure

import (
	"fmt"
	"runtime/debug"
	"sync/atomic"

	"github.com/go-logr/logr"
	"github.com/newcloudtechnologies/memlimiter/stats"
)

var _ Operator = (*operatorImpl)(nil)

// operatorImpl is the implementation of the Operator interface.
type operatorImpl struct {
	*throttler

	notificationChan      chan<- *stats.MemLimiterStats
	lastControlParameters atomic.Value
	logger                logr.Logger
}

// NewOperator constructs a new Operator.
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

// GetStats returns the current backpressure stats.
func (b *operatorImpl) GetStats() (*stats.BackpressureStats, error) {
	result := &stats.BackpressureStats{
		Throttling: b.getStats(),
	}

	lastControlParameters := b.lastControlParameters.Load()
	if lastControlParameters != nil {
		var ok bool

		result.ControlParameters, ok = lastControlParameters.(*stats.ControlParameters)
		if !ok {
			return nil, fmt.Errorf("invalid type cast (%T)", lastControlParameters)
		}
	}

	return result, nil
}

// SetControlParameters sets the control parameters.
func (b *operatorImpl) SetControlParameters(value *stats.ControlParameters) error {
	old := b.lastControlParameters.Swap(value)
	if old != nil {
		oldControlParameters, ok := old.(*stats.ControlParameters)
		if !ok {
			return fmt.Errorf("invalid type cast (%T)", old)
		}

		// If control parameters didn't change, we do nothing.
		if value.EqualsTo(oldControlParameters) {
			return nil
		}
	}

	// Set the share of the requests that have to be throttled.
	err := b.setThreshold(value.ThrottlingPercentage)
	if err != nil {
		return fmt.Errorf("throttler set threshold: %w", err)
	}

	// Tune GC pace.
	debug.SetGCPercent(value.GOGC)

	b.logger.Info("control parameters changed", value.ToKeysAndValues()...)

	// Notify client about statistics change.
	if b.notificationChan != nil {
		backpressureStats, err := b.GetStats()
		if err != nil {
			return fmt.Errorf("get stats: %w", err)
		}

		memLimiterStats := &stats.MemLimiterStats{
			Controller:   value.ControllerStats,
			Backpressure: backpressureStats,
		}

		// If client's not ready to read, omit it.
		select {
		case b.notificationChan <- memLimiterStats:
		default:
		}
	}

	return nil
}
