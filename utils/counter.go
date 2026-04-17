/*
 * Copyright (c) New Cloud Technologies, Ltd. 2013-2022.
 * Author: Vitaly Isaev <vitaly.isaev@myoffice.team>
 * License: https://github.com/newcloudtechnologies/memlimiter/blob/master/LICENSE
 */

package utils

import (
	"sync/atomic"
)

var _ Counter = (*counter)(nil)

type (
	// Counter is a thread-safe metrics counter.
	Counter interface {
		// Inc increments the counter by the given value.
		Inc(i int64)
		// Dec decrements the counter by the given value.
		Dec(i int64)
		// Count returns the current value of the counter.
		Count() int64
	}

	// counter is an atomic counter that may refer to a parent counter.
	// It must not be copied after first use because it contains atomic fields.
	counter struct {
		// value is the current value of the counter.
		value atomic.Int64
		// parent is the parent counter.
		// If parent is nil, the counter is a root in the hierarchy.
		parent Counter
	}
)

// NewCounter creates a counter referring to a parent counter.
// If parent is nil, the root in the hierarchy is created.
func NewCounter(parent Counter) Counter {
	return &counter{parent: parent}
}

// Inc increments the counter by the given value.
func (c *counter) Inc(i int64) {
	if c.parent != nil {
		c.parent.Inc(i)
	}

	c.value.Add(i)
}

// Dec decrements the counter by the given value.
func (c *counter) Dec(i int64) {
	if c.parent != nil {
		c.parent.Dec(i)
	}

	c.value.Add(-i)
}

// Count returns the current value of the counter.
func (c *counter) Count() int64 {
	return c.value.Load()
}
