/*
 * Copyright (c) New Cloud Technologies, Ltd. 2013-2022.
 * Author: Vitaly Isaev <vitaly.isaev@myoffice.team>
 * License: https://github.com/newcloudtechnologies/memlimiter/blob/master/LICENSE
 */

package perf

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"github.com/newcloudtechnologies/memlimiter/utils"
	"github.com/newcloudtechnologies/memlimiter/utils/config/prepare"
	"golang.org/x/time/rate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/durationpb"

	"github.com/newcloudtechnologies/memlimiter/test/allocator/schema"
	"github.com/newcloudtechnologies/memlimiter/utils/breaker"
)

// Client - client for performance testing.
type Client struct {
	startTime        time.Time
	client           schema.AllocatorClient
	requestsInFlight utils.Counter[int64]
	grpcConn         *grpc.ClientConn
	breaker          *breaker.Breaker
	cfg              *Config
	logger           logr.Logger
}

// NewClient creates new client for performance tests.
func NewClient(logger logr.Logger, cfg *Config) (*Client, error) {
	if err := prepare.Prepare(cfg); err != nil {
		return nil, fmt.Errorf("configs prepare: %w", err)
	}

	grpcConn, err := grpc.NewClient(cfg.Endpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("dial error: %w", err)
	}

	client := schema.NewAllocatorClient(grpcConn)

	return &Client{
		grpcConn:         grpcConn,
		logger:           logger,
		client:           client,
		startTime:        time.Now(),
		cfg:              cfg,
		requestsInFlight: utils.NewInt64Counter(nil),
		breaker:          breaker.NewBreaker(),
	}, nil
}

// Run starts load session.
func (p *Client) Run() error {
	err := p.breaker.Inc()
	if err != nil {
		return fmt.Errorf("breaker inc: %w", err)
	}

	defer p.breaker.Dec()

	monitoringTicker := time.NewTicker(time.Second)
	defer monitoringTicker.Stop()

	timer := time.NewTimer(p.cfg.LoadDuration.Duration)
	defer timer.Stop()

	limiter := rate.NewLimiter(p.cfg.RPS, 1)

	// Single threaded for simplicity.
	for {
		// Wait till limiter allows to fire a request.
		err := limiter.Wait(p.breaker)
		if err != nil {
			return fmt.Errorf("limiter wait: %w", err)
		}

		// Increment request copies.
		err = p.breaker.Inc()
		if err != nil {
			return fmt.Errorf("breaker inc: %w", err)
		}

		go p.makeRequest()

		select {
		case <-monitoringTicker.C:
			// Print progress periodically.
			p.printProgress()
		case <-timer.C:
			// Terminate load.
			return nil
		default:
		}
	}
}

// Quit terminates perf client gracefully.
func (p *Client) Quit() {
	p.breaker.ShutdownAndWait()

	err := p.grpcConn.Close()
	if err != nil {
		p.logger.Error(err, "gprc connection close")
	}
}

// makeRequest makes a request to the allocator server.
func (p *Client) makeRequest() {
	defer p.breaker.Dec()

	// Update in-flight request counter.
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
	if err != nil && p.breaker.IsOperational() {
		p.logger.Error(err, "make allocation request")
	}
}

// printProgress prints the progress of the load session.
func (p *Client) printProgress() {
	p.logger.Info(
		"progress",
		"elapsed_time", time.Since(p.startTime),
		"in_flight", p.requestsInFlight.Count(),
	)
}
