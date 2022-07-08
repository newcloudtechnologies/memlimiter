/*
 * Copyright (c) New Cloud Technologies, Ltd. 2013-2022.
 * Author: Vitaly Isaev <vitaly.isaev@myoffice.team>
 * License: https://github.com/newcloudtechnologies/memlimiter/blob/master/LICENSE
 */

package memlimiter

import (
	"github.com/go-logr/logr"
	"github.com/newcloudtechnologies/memlimiter/backpressure"
	"github.com/newcloudtechnologies/memlimiter/controller"
	"github.com/newcloudtechnologies/memlimiter/controller/nextgc"
	"github.com/newcloudtechnologies/memlimiter/middleware"
	"github.com/newcloudtechnologies/memlimiter/stats"
	"github.com/newcloudtechnologies/memlimiter/utils/config/prepare"
	"github.com/pkg/errors"
)

var _ Service = (*serviceImpl)(nil)

type serviceImpl struct {
	middleware           middleware.Middleware
	backpressureOperator backpressure.Operator
	statsSubscription    stats.ServiceStatsSubscription
	controller           controller.Controller
	logger               logr.Logger
}

func (s *serviceImpl) Middleware() middleware.Middleware { return s.middleware }

func (s *serviceImpl) GetStats() (*stats.MemLimiterStats, error) {
	controllerStats, err := s.controller.GetStats()
	if err != nil {
		return nil, errors.Wrap(err, "controller tracker")
	}

	backpressureStats, err := s.backpressureOperator.GetStats()
	if err != nil {
		return nil, errors.Wrap(err, "backpressure tracker")
	}

	return &stats.MemLimiterStats{
		Controller:   controllerStats,
		Backpressure: backpressureStats,
	}, nil
}

func (s *serviceImpl) Quit() {
	s.controller.Quit()
	s.statsSubscription.Quit()
}

// newServiceImpl - main entrypoint for MemLimiter.
func newServiceImpl(
	logger logr.Logger,
	cfg *Config,
	statsSubscription stats.ServiceStatsSubscription,
	backpressureOperator backpressure.Operator,
) (Service, error) {
	if err := prepare.Prepare(cfg); err != nil {
		return nil, errors.Wrap(err, "prepare config")
	}

	if statsSubscription == nil {
		return nil, errors.New("nil tracker subscription passed")
	}

	c, err := nextgc.NewControllerFromConfig(
		logger,
		cfg.ControllerNextGC,
		statsSubscription,
		backpressureOperator,
	)

	if err != nil {
		return nil, errors.Wrap(err, "new controller from config")
	}

	return &serviceImpl{
		middleware:           middleware.NewMiddleware(logger, backpressureOperator),
		backpressureOperator: backpressureOperator,
		statsSubscription:    statsSubscription,
		controller:           c,
		logger:               logger,
	}, nil
}
