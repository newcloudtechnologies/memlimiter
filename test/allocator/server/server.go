package server

import (
	"context"
	"math/rand"
	"net"
	"time"

	"github.com/go-logr/logr"
	"github.com/newcloudtechnologies/memlimiter"
	"github.com/newcloudtechnologies/memlimiter/stats"
	"github.com/newcloudtechnologies/memlimiter/utils"
	"github.com/newcloudtechnologies/memlimiter/utils/config/prepare"
	"google.golang.org/grpc"

	"github.com/newcloudtechnologies/memlimiter/test/allocator/schema"
	"github.com/pkg/errors"
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
	// MemLimiter returns underlying memlimiter implementation. Only for testing purposes.
	MemLimiter() memlimiter.Service
}

var _ Server = (*serverImpl)(nil)

type serverImpl struct {
	schema.UnimplementedAllocatorServer
	ml         memlimiter.Service
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

	// какая-то имитация работы, чтоб компилятор не оптимизировал массив
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

func (srv *serverImpl) MemLimiter() memlimiter.Service { return srv.ml }

func (srv *serverImpl) Quit() {
	srv.logger.Info("terminating server")
	srv.grpcServer.Stop()
}

// NewAllocatorServer - server constructor.
func NewAllocatorServer(logger logr.Logger, cfg *Config, options ...grpc.ServerOption) (Server, error) {
	if err := prepare.Prepare(cfg); err != nil {
		return nil, errors.Wrap(err, "configs prepare")
	}

	ml, err := memlimiter.NewServiceFromConfig(
		logger,
		cfg.MemLimiter,
		utils.NewUngracefulApplicationTerminator(logger),
		stats.NewSubscriptionDefault(time.Second),
	)

	if err != nil {
		return nil, errors.Wrap(err, "new memlimiter from config")
	}

	options = append(options,
		grpc.UnaryInterceptor(ml.Middleware().GRPC().MakeUnaryServerInterceptor()),
		grpc.StreamInterceptor(ml.Middleware().GRPC().MakeStreamServerInterceptor()),
	)

	srv := &serverImpl{
		logger:     logger,
		cfg:        cfg,
		ml:         ml,
		grpcServer: grpc.NewServer(options...),
	}

	schema.RegisterAllocatorServer(srv.grpcServer, srv)

	return srv, nil
}
