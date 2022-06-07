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

// Server - интерфейс сервера.
type Server interface {
	schema.AllocatorServer
	// Run запускает в работу сервис. Блокирующий вызов.
	Run() error
	// Quit корректное завершение работы сервера.
	Quit()
	// GRPCServer returns underlying server implementation. Only for test purposes.
	GRPCServer() *grpc.Server
	// MemLimiter returns underlying memlimiter implementation. Only for test purposes.
	MemLimiter() memlimiter.Service
}

var _ Server = (*serverImpl)(nil)

type serverImpl struct {
	schema.UnimplementedAllocatorServer
	cfg        *Config
	logger     logr.Logger
	grpcServer *grpc.Server
	ml         memlimiter.Service
}

func (srv *serverImpl) MakeAllocation(_ context.Context, request *schema.MakeAllocationRequest) (*schema.MakeAllocationResponse, error) {
	var array []byte

	// аллоцируем массив
	if request.Size != 0 {
		array = make([]byte, int(request.Size))
		//nolint:gosec
		if _, err := rand.Read(array); err != nil {
			return nil, errors.Wrap(err, "rand read")
		}
	}

	// ждём определённое время, чтобы он побыл в оперативной памяти - это имитация бизнес-логики
	duration := request.Duration.AsDuration()
	if duration != 0 {
		time.Sleep(duration)
	}

	// какая-то имитация работы, чтоб компилятор не оптимизировал массив
	x := uint64(0)
	for i := 0; i < len(array); i++ {
		x += uint64(array[i])
	}

	return &schema.MakeAllocationResponse{Value: x}, nil
}

func (srv *serverImpl) Run() error {
	endpoint := srv.cfg.Server.ListenEndpoint

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

// NewAllocatorServer - конструктор сервера.
func NewAllocatorServer(logger logr.Logger, cfg *Config, options ...grpc.ServerOption) (Server, error) {
	if err := prepare.Prepare(cfg); err != nil {
		return nil, errors.Wrap(err, "configs prepare")
	}

	ml, err := memlimiter.NewServiceFromConfig(
		logger,
		cfg.MemLimiter,
		utils.NewUngracefulApplicationTerminator(logger),
		stats.NewSubscriptionDefault(time.Second),
		nil,
	)

	if err != nil {
		return nil, errors.Wrap(err, "new memlimiter from config")
	}

	options = append(options,
		grpc.UnaryInterceptor(ml.MakeUnaryServerInterceptor()),
		grpc.StreamInterceptor(ml.MakeStreamServerInterceptor()),
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
