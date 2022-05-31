package memlimiter

import (
	"context"
	"sync/atomic"

	"github.com/go-logr/logr"
	"github.com/newcloudtechnologies/memlimiter/stats"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/newcloudtechnologies/memlimiter/backpressure"
	"github.com/newcloudtechnologies/memlimiter/controller"
	"github.com/newcloudtechnologies/memlimiter/controller/nextgc"
	"github.com/newcloudtechnologies/memlimiter/utils"
	"github.com/pkg/errors"
)

// memLimiterImpl - система управления бюджетом оперативной памяти.
type memLimiterImpl struct {
	controller            atomic.Value // в связи с ленивой инициализацией
	backpressureOperator  backpressure.Operator
	consumptionReporter   utils.ConsumptionReporter
	applicationTerminator utils.ApplicationTerminator
	statsSubscription     stats.Subscription
	logger                logr.Logger
	cfg                   *Config
}

func (ml *memLimiterImpl) Init(statsSubscription stats.Subscription) error {
	if c := ml.controller.Load(); c != nil {
		return errors.New("memlimiter is already initialized")
	}

	ml.statsSubscription = statsSubscription

	// NOTE: здесь должен появиться switch по типам при появлении иных типов контроллеров
	c := nextgc.NewControllerFromConfig(
		ml.logger,
		ml.cfg.ControllerNextGC,
		ml.statsSubscription,
		ml.consumptionReporter,
		ml.backpressureOperator,
		ml.applicationTerminator,
	)
	ml.controller.Store(c)

	return nil
}

func (ml *memLimiterImpl) GetStats() (*stats.Memlimiter, error) {
	c := ml.controller.Load()
	if c == nil {
		return nil, errors.New("memlimiter is not initialized yet")
	}

	controllerStats, err := c.(controller.Controller).GetStats()
	if err != nil {
		return nil, errors.Wrap(err, "controller get stats")
	}

	backpressureStats := ml.backpressureOperator.GetStats()

	return &stats.Memlimiter{
		Controller:   controllerStats,
		Backpressure: backpressureStats,
	}, nil
}

// MakeUnaryServerInterceptor - унарный интерсептор, выполняющий подавление запросов.
func (ml *memLimiterImpl) MakeUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		allowed := ml.backpressureOperator.AllowRequest()
		if allowed {
			return handler(ctx, req)
		}

		logger, err := logr.FromContext(ctx)
		if err != nil {
			logger = ml.logger
		}

		logger.Info("request has been throttled")

		return nil, status.Error(codes.ResourceExhausted, "request has been throttled")
	}
}

// MakeStreamServerInterceptor - стримовый интерсептор, выполняющий подавление запросов.
func (ml *memLimiterImpl) MakeStreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		allowed := ml.backpressureOperator.AllowRequest()
		if allowed {
			return handler(srv, ss)
		}

		logger, err := logr.FromContext(ss.Context())
		if err != nil {
			logger = ml.logger
		}

		logger.Info("request has been throttled")

		return status.Error(codes.ResourceExhausted, "request has been throttled")
	}
}

// Quit корректно завершает работу.
func (ml *memLimiterImpl) Quit() {
	c := ml.controller.Load()
	if c == nil {
		return
	}

	c.(controller.Controller).Quit()
	ml.statsSubscription.Quit()
}

func newMemLimiterDefault(
	logger logr.Logger, // обязательный
	cfg *Config, // обязательный
	applicationTerminator utils.ApplicationTerminator, // обязательный
	consumptionReporter utils.ConsumptionReporter, // опциональный
) (MemLimiter, error) {
	if applicationTerminator == nil {
		return nil, errors.New("nil application terminator passed")
	}

	backpressureOperator := backpressure.NewOperator(logger)

	return &memLimiterImpl{
		cfg:                   cfg,
		consumptionReporter:   consumptionReporter,
		backpressureOperator:  backpressureOperator,
		applicationTerminator: applicationTerminator,
		logger:                logger,
	}, nil
}
