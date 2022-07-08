/*
 * Copyright (c) New Cloud Technologies, Ltd. 2013-2022.
 * Author: Vitaly Isaev <vitaly.isaev@myoffice.team>
 * License: https://github.com/newcloudtechnologies/memlimiter/blob/master/LICENSE
 */

package memlimiter

import (
	"github.com/newcloudtechnologies/memlimiter/backpressure"
	"github.com/newcloudtechnologies/memlimiter/stats"
)

// Option - MemLimiter constructor options.
type Option interface {
	anchor()
}

type backpressureOperatorOption struct {
	val backpressure.Operator
}

func (b *backpressureOperatorOption) anchor() {}

// WithBackpressureOperator allows client to provide customized backpressure.Operator;
// that's especially useful when implementing backpressure logic on the application side.
func WithBackpressureOperator(val backpressure.Operator) Option {
	return &backpressureOperatorOption{val: val}
}

type serviceStatsSubscriptionOption struct {
	val stats.ServiceStatsSubscription
}

func (s serviceStatsSubscriptionOption) anchor() {
}

// WithServiceStatsSubscription allows client to provide own implementation of service stats subscription.
func WithServiceStatsSubscription(val stats.ServiceStatsSubscription) Option {
	return &serviceStatsSubscriptionOption{val: val}
}
