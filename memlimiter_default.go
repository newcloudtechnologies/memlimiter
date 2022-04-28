package memlimiter

import (
	"context"
	"sync/atomic"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	servus_stats "gitlab.stageoffice.ru/UCS-COMMON/schemagen-go/v41/servus/stats/v1"

	"github.com/pkg/errors"
	"gitlab.stageoffice.ru/UCS-COMMON/gaben"
	"gitlab.stageoffice.ru/UCS-PLATFORM/servus"
	"gitlab.stageoffice.ru/UCS-PLATFORM/servus/stats/aggregate"

	"github.com/newcloudtechnologies/memlimiter/backpressure"
	"github.com/newcloudtechnologies/memlimiter/controller"
	"github.com/newcloudtechnologies/memlimiter/controller/nextgc"
	"github.com/newcloudtechnologies/memlimiter/utils"
)

// memLimiterImpl - система управления бюджетом оперативной памяти.
type memLimiterImpl struct {
	controller            atomic.Value // в связи с ленивой инициализацией
	backpressureOperator  backpressure.Operator
	consumptionReporter   utils.ConsumptionReporter
	applicationTerminator utils.ApplicationTerminator
	statsSubscription     aggregate.Subscription
	logger                gaben.Logger
	cfg                   *Config
}

func (ml *memLimiterImpl) Init(ss servus.Servus) error {
	if c := ml.controller.Load(); c != nil {
		return errors.New("memlimiter is already initialized")
	}

	ml.statsSubscription = ss.AggregatedStats().Subscribe()

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

func (ml *memLimiterImpl) getStats() (*servus_stats.GoMemLimiterStats, error) {
	c := ml.controller.Load()
	if c == nil {
		return nil, errors.New("memlimiter is not initialized yet")
	}

	controllerStats, err := c.(controller.Controller).GetStats()
	if err != nil {
		return nil, errors.Wrap(err, "controller get stats")
	}

	backpressureStats := ml.backpressureOperator.GetStats()

	return &servus_stats.GoMemLimiterStats{
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

		logger, ok := gaben.FromContext(ctx)
		if !ok {
			logger = ml.logger
		}

		logger.Warning("request has been throttled")

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

		logger, ok := gaben.FromContext(ss.Context())
		if !ok {
			logger = ml.logger
		}

		logger.Warning("request has been throttled")

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
	logger gaben.Logger, // обязательный
	cfg *Config, // обязательный
	applicationTerminator utils.ApplicationTerminator, // обязательный
	consumptionReporter utils.ConsumptionReporter, // опциональный
) (MemLimiter, error) {
	if logger == nil {
		return nil, errors.New("nil logger passed")
	}

	logger = gaben.Spawn(logger).With(gaben.String("subsystem", "memlimiter"))

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
