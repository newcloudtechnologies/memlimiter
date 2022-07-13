/*
 * Copyright (c) New Cloud Technologies, Ltd. 2013-2022.
 * Author: Vitaly Isaev <vitaly.isaev@myoffice.team>
 * License: https://github.com/newcloudtechnologies/memlimiter/blob/master/LICENSE
 */

package memlimiter

import (
	"time"

	"github.com/go-logr/logr"
	"github.com/newcloudtechnologies/memlimiter/backpressure"
	"github.com/newcloudtechnologies/memlimiter/stats"
)

// NewServiceFromConfig - main entrypoint for MemLimiter.
func NewServiceFromConfig(
	logger logr.Logger,
	cfg *Config,
	options ...Option,
) (Service, error) {
	var (
		serviceStatsSubscription stats.ServiceStatsSubscription
		backpressureOperator     backpressure.Operator
	)

	for _, op := range options {
		switch t := (op).(type) {
		case *serviceStatsSubscriptionOption:
			serviceStatsSubscription = t.val
		case *backpressureOperatorOption:
			backpressureOperator = t.val
		}
	}

	// make defaults
	if serviceStatsSubscription == nil {
		serviceStatsSubscription = stats.NewSubscriptionDefault(logger, time.Second)
	}

	if backpressureOperator == nil {
		backpressureOperator = backpressure.NewOperator(logger)
	}

	if cfg == nil {
		return newServiceStub(serviceStatsSubscription), nil
	}

	return newServiceImpl(logger, cfg, serviceStatsSubscription, backpressureOperator)
}
