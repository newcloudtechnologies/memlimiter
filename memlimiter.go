package memlimiter

import (
	"github.com/go-logr/logr"
	"github.com/newcloudtechnologies/memlimiter/backpressure"
	"github.com/newcloudtechnologies/memlimiter/controller"
	"github.com/newcloudtechnologies/memlimiter/controller/nextgc"
	"github.com/newcloudtechnologies/memlimiter/middleware"
	"github.com/newcloudtechnologies/memlimiter/stats"
	"github.com/newcloudtechnologies/memlimiter/utils"
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
		return nil, errors.Wrap(err, "controller get stats")
	}

	backpressureStats := s.backpressureOperator.GetStats()

	return &stats.MemLimiterStats{
		Controller:   controllerStats,
		Backpressure: backpressureStats,
	}, nil
}

func (s *serviceImpl) Quit() {
	s.controller.Quit()
	s.statsSubscription.Quit()
}

// NewServiceFromConfig - main entrypoint for MemLimiter.
func NewServiceFromConfig(
	logger logr.Logger,
	cfg *Config,
	applicationTerminator utils.ApplicationTerminator,
	statsSubscription stats.ServiceStatsSubscription,
) (Service, error) {
	if err := prepare.Prepare(cfg); err != nil {
		return nil, errors.Wrap(err, "prepare config")
	}

	if applicationTerminator == nil {
		return nil, errors.New("nil application terminator passed")
	}

	if statsSubscription == nil {
		return nil, errors.New("nil stats subscription passed")
	}

	backpressureOperator := backpressure.NewOperator(logger)

	c := nextgc.NewControllerFromConfig(
		logger,
		cfg.ControllerNextGC,
		statsSubscription,
		backpressureOperator,
		applicationTerminator,
	)

	return &serviceImpl{
		middleware:           middleware.NewMiddleware(logger, backpressureOperator),
		backpressureOperator: backpressureOperator,
		statsSubscription:    statsSubscription,
		controller:           c,
		logger:               logger,
	}, nil
}
