package perf

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	"github.com/newcloudtechnologies/memlimiter/utils/config/prepare"
	"github.com/rcrowley/go-metrics"
	"golang.org/x/time/rate"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/durationpb"

	"github.com/newcloudtechnologies/memlimiter/test/allocator/schema"
	"github.com/newcloudtechnologies/memlimiter/utils/breaker"
	"github.com/pkg/errors"
)

// Client - нагрузочный клиент.
type Client struct {
	startTime        time.Time
	logger           logr.Logger
	grpcConn         *grpc.ClientConn
	client           schema.AllocatorClient
	breaker          *breaker.Breaker
	requestsInFlight metrics.Counter
	cfg              *Config
}

// Run запускает нагрузочную сессию.
func (p *Client) Run() error {
	if err := p.breaker.Inc(); err != nil {
		return errors.Wrap(err, "breaker inc")
	}

	defer p.breaker.Dec()

	monitoringTicker := time.NewTicker(time.Second)
	defer monitoringTicker.Stop()

	timer := time.NewTimer(p.cfg.LoadDuration.Duration)
	defer timer.Stop()

	limiter := rate.NewLimiter(p.cfg.RPS, 1)

	for {
		// ожидаем, пока лимитер разрешит выполнять запрос
		if err := limiter.Wait(p.breaker); err != nil {
			return errors.Wrap(err, "limiter wait")
		}

		// запрос
		if err := p.breaker.Inc(); err != nil {
			return errors.Wrap(err, "breaker inc")
		}

		go p.makeRequest()

		select {
		case <-monitoringTicker.C:
			// периодическая печать прогресса
			p.printProgress()
		case <-timer.C:
			// завершение нагрузки
			return nil
		default:
		}
	}
}

func (p *Client) makeRequest() {
	defer p.breaker.Dec()

	// обновление счётчика запросов в полете
	p.requestsInFlight.Inc(1)
	defer p.requestsInFlight.Dec(1)

	ctx, cancel := context.WithTimeout(p.breaker, p.cfg.RequestTimeout.Duration)
	defer cancel()

	request := &schema.MakeAllocationRequest{
		Size: p.cfg.AllocationSize.Value,
	}

	if p.cfg.PauseDuration.Duration != 0 {
		request.Duration = durationpb.New(p.cfg.PauseDuration.Duration)
	}

	_, err := p.client.MakeAllocation(ctx, request)
	if err != nil {
		p.logger.Error(err, "make allocation request")
	}
}

func (p *Client) printProgress() {
	p.logger.Info(
		"progress",
		"elapsed_time", time.Since(p.startTime),
		"in_flight", p.requestsInFlight.Count(),
	)
}

// Quit корректно завершает работу нагрузчика.
func (p *Client) Quit() {
	p.breaker.ShutdownAndWait()

	if err := p.grpcConn.Close(); err != nil {
		p.logger.Error(err, "gprc connection close")
	}
}

// NewClient создаёт нагрузочный клиент.
func NewClient(cfg *Config) (*Client, error) {
	if err := prepare.Prepare(cfg); err != nil {
		return nil, errors.Wrap(err, "configs prepare")
	}

	// FIXME:
	/*
		logger, err := gaben.FromConfig(cfg.Logging)
		if err != nil {
			return nil, errors.Wrap(err, "gaben from config")
		}
	*/

	grpcConn, err := grpc.Dial(cfg.Endpoint, grpc.WithInsecure())
	if err != nil {
		return nil, errors.Wrap(err, "dial error")
	}

	client := schema.NewAllocatorClient(grpcConn)

	return &Client{
		grpcConn:         grpcConn,
		logger:           logr.Logger{}, // FIXME
		client:           client,
		startTime:        time.Now(),
		cfg:              cfg,
		requestsInFlight: metrics.NewCounter(),
		breaker:          breaker.NewBreaker(),
	}, nil
}
