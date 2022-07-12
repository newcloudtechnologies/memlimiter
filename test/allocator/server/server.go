/*
 * Copyright (c) New Cloud Technologies, Ltd. 2013-2022.
 * Author: Vitaly Isaev <vitaly.isaev@myoffice.team>
 * License: https://github.com/newcloudtechnologies/memlimiter/blob/master/LICENSE
 */

package server

import (
	"context"
	"math/rand"
	"net"
	"time"

	"github.com/go-logr/logr"
	"github.com/newcloudtechnologies/memlimiter"
	"github.com/newcloudtechnologies/memlimiter/test/allocator/schema"
	"github.com/newcloudtechnologies/memlimiter/test/allocator/tracker"
	"github.com/newcloudtechnologies/memlimiter/utils/config/prepare"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

// Server represents Allocator service interface.
type Server interface {
	schema.AllocatorServer
	// Run starts service (a blocking call).
	Run() error
	// Quit terminates service gracefully.
	Quit()
	// GRPCServer returns underlying server implementation. Only for testing purposes.
	GRPCServer() *grpc.Server
	// MemLimiter returns internal MemLimiter object. Only for testing purposes.
	MemLimiter() memlimiter.Service
	// Tracker returns statistics tracker. Only for testing purposes.
	Tracker() *tracker.Tracker
}

var _ Server = (*serverImpl)(nil)

type serverImpl struct {
	schema.UnimplementedAllocatorServer
	memLimiter memlimiter.Service
	tracker    *tracker.Tracker
	cfg        *Config
	grpcServer *grpc.Server
	logger     logr.Logger
}

func (srv *serverImpl) MakeAllocation(_ context.Context, request *schema.MakeAllocationRequest) (*schema.MakeAllocationResponse, error) {
	var slice []byte

	// allocate slice
	if request.Size != 0 {
		slice = make([]byte, int(request.Size))
		//nolint:gosec
		if _, err := rand.Read(slice); err != nil {
			return nil, errors.Wrap(err, "rand read")
		}
	}

	// Wait some time to make slice reside in the RSS (otherwise it could be immediately collected by GC).
	// This is a trivial imitation of a real-world service business logic.
	duration := request.Duration.AsDuration()
	if duration != 0 {
		time.Sleep(duration)
	}

	// Imitate some work with slice to prevent compiler from optimizing out the slice.
	x := uint64(0)
	for i := 0; i < len(slice); i++ {
		x += uint64(slice[i])
	}

	return &schema.MakeAllocationResponse{Value: x}, nil
}

func (srv *serverImpl) Run() error {
	endpoint := srv.cfg.ListenEndpoint

	listener, err := net.Listen("tcp", endpoint)
	if err != nil {
		return errors.Wrap(err, "net listen")
	}

	srv.logger.Info("starting listening", "endpoint", endpoint)

	if err = srv.grpcServer.Serve(listener); err != nil {
		return errors.Wrap(err, "grpc server serve")
	}

	return nil
}

func (srv *serverImpl) GRPCServer() *grpc.Server { return srv.grpcServer }

func (srv *serverImpl) MemLimiter() memlimiter.Service { return srv.memLimiter }

func (srv *serverImpl) Tracker() *tracker.Tracker { return srv.tracker }

func (srv *serverImpl) Quit() {
	srv.logger.Info("terminating server")
	srv.grpcServer.Stop()
}

// NewServer - server constructor.
func NewServer(logger logr.Logger, cfg *Config, options ...grpc.ServerOption) (Server, error) {
	if err := prepare.Prepare(cfg); err != nil {
		return nil, errors.Wrap(err, "configs prepare")
	}

	memLimiter, err := memlimiter.NewServiceFromConfig(logger, cfg.MemLimiter)
	if err != nil {
		return nil, errors.Wrap(err, "new MemLimiter from config")
	}

	tr, err := tracker.NewTrackerFromConfig(logger, cfg.Tracker, memLimiter)
	if err != nil {
		return nil, errors.Wrap(err, "new tracker from config")
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
