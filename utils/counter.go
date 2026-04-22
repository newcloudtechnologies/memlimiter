/*
 * Copyright (c) New Cloud Technologies, Ltd. 2013-2026.
 * Author: Vitaly Isaev <vitaly.isaev@myoffice.team>
 * License: https://github.com/newcloudtechnologies/memlimiter/blob/master/LICENSE
 */

package utils

import (
	"sync/atomic"
)

var (
	_ Counter[int64]  = (*counterInt64)(nil)
	_ Counter[uint64] = (*counterUint64)(nil)
)

// CounterValue is a supported counter numeric type.
// constraints.Integer is too wide here,
// we should use the narrowest constraint matching real implementation.
type CounterValue interface {
	~int64 | ~uint64
}

// Counter is a thread-safe metrics counter.
type Counter[T CounterValue] interface {
	// Inc increments the counter by the given value.
	Inc(i T)
	// Dec decrements the counter by the given value.
	Dec(i T)
	// Count returns the current value of the counter.
	Count() T
}

// counterInt64 is an atomic int64 counter that may refer to a parent counter.
// It must not be copied after first use because it contains atomic fields.
type counterInt64 struct {
	// parent is the parent counter.
	// If parent is nil, the counter is a root in the hierarchy.
	parent Counter[int64]
	// value is the current value of the counter.
	value atomic.Int64
}

// counterUint64 is an atomic uint64 counter that may refer to a parent counter.
// It must not be copied after first use because it contains atomic fields.
type counterUint64 struct {
	// parent is the parent counter.
	// If parent is nil, the counter is a root in the hierarchy.
	parent Counter[uint64]
	// value is the current value of the counter.
	value atomic.Uint64
}

// NewInt64Counter creates an int64 counter referring to a parent counter.
// If parent is nil, the root in the hierarchy is created.
func NewInt64Counter(parent Counter[int64]) Counter[int64] {
	return &counterInt64{
		parent: parent,
	}
}

// NewUint64Counter creates a uint64 counter referring to a parent counter.
// If parent is nil, the root in the hierarchy is created.
func NewUint64Counter(parent Counter[uint64]) Counter[uint64] {
	return &counterUint64{
		parent: parent,
	}
}

// Inc increments the counter by the given value.
func (c *counterInt64) Inc(i int64) {
	if c.parent != nil {
		c.parent.Inc(i)
	}

	c.value.Add(i)
}

// Dec decrements the counter by the given value.
func (c *counterInt64) Dec(i int64) {
	if c.parent != nil {
		c.parent.Dec(i)
	}

	c.value.Add(-i)
}

// Count returns the current value of the counter.
func (c *counterInt64) Count() int64 {
	return c.value.Load()
}

// Inc increments the counter by the given value.
func (c *counterUint64) Inc(i uint64) {
	if c.parent != nil {
		c.parent.Inc(i)
	}

	c.value.Add(i)
}

// Dec decrements the counter by the given value.
func (c *counterUint64) Dec(i uint64) {
	if c.parent != nil {
		c.parent.Dec(i)
	}

	if i == 0 {
		return
	}

	// Subtract using two's complement arithmetic.
	c.value.Add(^(i - 1))
}

// Count returns the current value of the counter.
func (c *counterUint64) Count() uint64 {
	return c.value.Load()
}
