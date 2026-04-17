/*
 * Copyright (c) New Cloud Technologies, Ltd. 2013-2022.
 * Author: Vitaly Isaev <vitaly.isaev@myoffice.team>
 * License: https://github.com/newcloudtechnologies/memlimiter/blob/master/LICENSE
 */

package memlimiter

import (
	"errors"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/newcloudtechnologies/memlimiter/backpressure"
	"github.com/newcloudtechnologies/memlimiter/controller"
	"github.com/newcloudtechnologies/memlimiter/controller/nextgc"
	"github.com/newcloudtechnologies/memlimiter/middleware"
	"github.com/newcloudtechnologies/memlimiter/stats"
	"github.com/newcloudtechnologies/memlimiter/utils/config/prepare"
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
		return nil, fmt.Errorf("controller tracker: %w", err)
	}

	backpressureStats, err := s.backpressureOperator.GetStats()
	if err != nil {
		return nil, fmt.Errorf("backpressure tracker: %w", err)
	}

	return &stats.MemLimiterStats{
		Controller:   controllerStats,
		Backpressure: backpressureStats,
	}, nil
}

func (s *serviceImpl) Quit() {
	s.logger.Info("terminating MemLimiter service")
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
		return nil, fmt.Errorf("prepare config: %w", err)
	}

	if statsSubscription == nil {
		return nil, errors.New("nil tracker subscription passed")
	}

	logger.Info("starting MemLimiter service")

	c, err := nextgc.NewControllerFromConfig(
		logger,
		cfg.ControllerNextGC,
		statsSubscription,
		backpressureOperator,
	)

	if err != nil {
		return nil, fmt.Errorf("new controller from config: %w", err)
	}

	return &serviceImpl{
		middleware:           middleware.NewMiddleware(logger, backpressureOperator),
		backpressureOperator: backpressureOperator,
		statsSubscription:    statsSubscription,
		controller:           c,
		logger:               logger,
	}, nil
}
