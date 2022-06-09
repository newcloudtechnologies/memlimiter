/*
 * Copyright (c) New Cloud Technologies, Ltd. 2013-2022.
 * Author: Vitaly Isaev <vitaly.isaev@myoffice.team>
 * License: https://github.com/newcloudtechnologies/memlimiter/blob/master/LICENSE
 */

package utils

import (
	go_metrics "github.com/rcrowley/go-metrics"
)

var _ Counter = (*childCounter)(nil)

// Counter - thread-safe metrics counter.
type Counter interface {
	go_metrics.Counter
}

// childCounter allows to construct hierarchical counters.
type childCounter struct {
	Counter
	parent Counter
}

func (counter *childCounter) Dec(i int64) {
	counter.parent.Dec(i)
	counter.Counter.Dec(i)
}

func (counter *childCounter) Inc(i int64) {
	counter.parent.Inc(i)
	counter.Counter.Inc(i)
}

// NewCounter creates counter referring to parent counter.
// If parent is nil, the root in hierarchy is created.
func NewCounter(parent Counter) Counter {
	if parent == nil {
		return go_metrics.NewCounter()
	}

	return &childCounter{Counter: go_metrics.NewCounter(), parent: parent}
}
