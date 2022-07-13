/*
 * Copyright (c) New Cloud Technologies, Ltd. 2013-2022.
 * Author: Vitaly Isaev <vitaly.isaev@myoffice.team>
 * License: https://github.com/newcloudtechnologies/memlimiter/blob/master/LICENSE
 */

package backpressure

import (
	"github.com/newcloudtechnologies/memlimiter/stats"
)

// Option - backpressure operator options.
type Option interface {
	anchor()
}

type notificationsOption struct {
	val chan<- *stats.MemLimiterStats
}

func (o notificationsOption) anchor() {}

// WithNotificationsOption makes it possible for a client to implement application-specific
// backpressure logic. Client may subscribe for the operative MemLimiter telemetry and make
// own decisions (taking into account metrics like RSS etc.).
// This is inspired by the channel that was attached to SetMaxHeap function
// (see https://github.com/golang/proposal/blob/master/design/48409-soft-memory-limit.md#setmaxheap)
func WithNotificationsOption(notifications chan<- *stats.MemLimiterStats) Option {
	return &notificationsOption{
		val: notifications,
	}
}
