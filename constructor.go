/*
 * Copyright (c) New Cloud Technologies, Ltd. 2013-2022.
 * Author: Vitaly Isaev <vitaly.isaev@myoffice.team>
 * License: https://github.com/newcloudtechnologies/memlimiter/blob/master/LICENSE
 */

package memlimiter

import (
	"github.com/go-logr/logr"
	"github.com/newcloudtechnologies/memlimiter/stats"
	"github.com/newcloudtechnologies/memlimiter/utils"
)

// NewServiceFromConfig - main entrypoint for MemLimiter.
func NewServiceFromConfig(
	logger logr.Logger,
	cfg *Config,
	applicationTerminator utils.ApplicationTerminator,
	statsSubscription stats.ServiceStatsSubscription,
) (Service, error) {
	if cfg == nil {
		return newServiceStub(statsSubscription), nil
	}

	return newServiceImpl(logger, cfg, applicationTerminator, statsSubscription)
}
