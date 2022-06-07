package memlimiter

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/newcloudtechnologies/memlimiter/backpressure"
	"github.com/newcloudtechnologies/memlimiter/controller"
	"github.com/newcloudtechnologies/memlimiter/controller/nextgc"
	"github.com/newcloudtechnologies/memlimiter/stats"
	"github.com/newcloudtechnologies/memlimiter/utils"
	"github.com/newcloudtechnologies/memlimiter/utils/config/prepare"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ Service = (*serviceImpl)(nil)

// serviceImpl - система управления бюджетом оперативной памяти.
type serviceImpl struct {
	backpressureOperator backpressure.Operator
	statsSubscription    stats.Subscription
	controller           controller.Controller
	logger               logr.Logger
}

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

// MakeUnaryServerInterceptor - унарный интерсептор, выполняющий подавление запросов.
func (s *serviceImpl) MakeUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		allowed := s.backpressureOperator.AllowRequest()
		if allowed {
			return handler(ctx, req)
		}

		logger, err := logr.FromContext(ctx)
		if err != nil {
			logger = s.logger
		}

		logger.Info("request has been throttled")

		return nil, status.Error(codes.ResourceExhausted, "request has been throttled")
	}
}

// MakeStreamServerInterceptor - стримовый интерсептор, выполняющий подавление запросов.
func (s *serviceImpl) MakeStreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		allowed := s.backpressureOperator.AllowRequest()
		if allowed {
			return handler(srv, ss)
		}

		logger, err := logr.FromContext(ss.Context())
		if err != nil {
			logger = s.logger
		}

		logger.Info("request has been throttled")

		return status.Error(codes.ResourceExhausted, "request has been throttled")
	}
}

// Quit корректно завершает работу.
func (s *serviceImpl) Quit() {
	s.controller.Quit()
	s.statsSubscription.Quit()
}

func NewServiceFromConfig(
	logger logr.Logger, // обязательный
	cfg *Config, // обязательный
	applicationTerminator utils.ApplicationTerminator, // обязательный
	statsSubscription stats.Subscription, // mandatory
	consumptionReporter utils.ConsumptionReporter, // опциональный
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
		consumptionReporter,
		backpressureOperator,
		applicationTerminator,
	)

	return &serviceImpl{
		backpressureOperator: backpressureOperator,
		statsSubscription:    statsSubscription,
		controller:           c,
		logger:               logger,
	}, nil
}
