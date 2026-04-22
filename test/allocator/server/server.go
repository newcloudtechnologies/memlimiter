/*
 * Copyright (c) New Cloud Technologies, Ltd. 2013-2022.
 * Author: Vitaly Isaev <vitaly.isaev@myoffice.team>
 * License: https://github.com/newcloudtechnologies/memlimiter/blob/master/LICENSE
 */

package server

import (
	"context"
	"fmt"
	"math"
	"math/rand/v2"
	"net"
	"time"

	"github.com/go-logr/logr"
	"github.com/newcloudtechnologies/memlimiter"
	"github.com/newcloudtechnologies/memlimiter/test/allocator/schema"
	"github.com/newcloudtechnologies/memlimiter/test/allocator/tracker"
	"github.com/newcloudtechnologies/memlimiter/utils/config/prepare"
	"google.golang.org/grpc"
)

// Server is the interface for the Allocator service.
type Server interface {
	schema.AllocatorServer
	// Run starts the server (a blocking call).
	Run() error
	// Quit terminates the server gracefully.
	Quit()
	// GRPCServer returns the underlying server implementation. Only for testing purposes.
	GRPCServer() *grpc.Server
	// MemLimiter returns the internal MemLimiter object. Only for testing purposes.
	MemLimiter() memlimiter.Service
	// Tracker returns the statistics tracker. Only for testing purposes.
	Tracker() *tracker.Tracker
}

var _ Server = (*serverImpl)(nil)

// serverImpl is the implementation of the Server interface.
type serverImpl struct {
	schema.UnimplementedAllocatorServer

	memLimiter memlimiter.Service
	tracker    *tracker.Tracker
	cfg        *Config
	grpcServer *grpc.Server
	logger     logr.Logger
}

// NewServer constructs a new server.
func NewServer(logger logr.Logger, cfg *Config, options ...grpc.ServerOption) (Server, error) {
	if err := prepare.Prepare(cfg); err != nil {
		return nil, fmt.Errorf("configs prepare: %w", err)
	}

	memLimiter, err := memlimiter.NewServiceFromConfig(logger, cfg.MemLimiter)
	if err != nil {
		return nil, fmt.Errorf("new MemLimiter from config: %w", err)
	}

	tr, err := tracker.NewTrackerFromConfig(logger, cfg.Tracker, memLimiter)
	if err != nil {
		return nil, fmt.Errorf("new tracker from config: %w", err)
	}

	if cfg.MemLimiter != nil {
		options = append(options,
			grpc.UnaryInterceptor(memLimiter.Middleware().GRPC().MakeUnaryServerInterceptor()),
			grpc.StreamInterceptor(memLimiter.Middleware().GRPC().MakeStreamServerInterceptor()),
		)
	}

	srv := &serverImpl{
		logger:     logger,
		cfg:        cfg,
		memLimiter: memLimiter,
		grpcServer: grpc.NewServer(options...),
		tracker:    tr,
	}

	schema.RegisterAllocatorServer(srv.grpcServer, srv)

	return srv, nil
}

// MakeAllocation makes an allocation.
func (srv *serverImpl) MakeAllocation(_ context.Context, request *schema.MakeAllocationRequest) (*schema.MakeAllocationResponse, error) {
	var slice []byte

	// Allocate slice.
	allocationSize := request.GetSize()
	if allocationSize != 0 {
		if allocationSize > uint64(math.MaxInt) {
			return nil, fmt.Errorf("allocation size is too large: %d", allocationSize)
		}

		slice = make([]byte, int(allocationSize))
		//nolint:gosec // Non-cryptographic RNG is enough for load-testing payload generation.
		for i := range slice {
			slice[i] = byte(rand.Uint64())
		}
	}

	// Wait some time to make slice reside in the RSS (otherwise it could be immediately collected by GC).
	// This is a trivial imitation of a real-world service business logic.
	duration := request.GetDuration().AsDuration()
	if duration != 0 {
		time.Sleep(duration)
	}

	// Imitate some work with slice to prevent compiler from optimizing out the slice.
	var x uint64
	for i := range slice {
		x += uint64(slice[i])
	}

	return &schema.MakeAllocationResponse{Value: x}, nil
}

// Run starts the server.
func (srv *serverImpl) Run() error {
	endpoint := srv.cfg.ListenEndpoint

	listenConfig := net.ListenConfig{}

	listener, err := listenConfig.Listen(context.Background(), "tcp", endpoint)
	if err != nil {
		return fmt.Errorf("net listen: %w", err)
	}

	srv.logger.Info("starting listening", "endpoint", endpoint)

	if err = srv.grpcServer.Serve(listener); err != nil {
		return fmt.Errorf("grpc server serve: %w", err)
	}

	return nil
}

// GRPCServer returns the underlying server implementation.
func (srv *serverImpl) GRPCServer() *grpc.Server { return srv.grpcServer }

// MemLimiter returns the internal MemLimiter object.
func (srv *serverImpl) MemLimiter() memlimiter.Service { return srv.memLimiter }

// Tracker returns the statistics tracker.
func (srv *serverImpl) Tracker() *tracker.Tracker { return srv.tracker }

// Quit terminates the server gracefully.
func (srv *serverImpl) Quit() {
	srv.logger.Info("terminating server")
	srv.grpcServer.Stop()
	srv.memLimiter.Quit()
}
