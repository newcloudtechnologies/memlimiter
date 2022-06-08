package middleware

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/newcloudtechnologies/memlimiter/backpressure"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GRPC provides server-side interceptors that must be used
// at the time of grpcImpl server construction.
type GRPC interface {
	// MakeUnaryServerInterceptor returns unary server interceptor.
	MakeUnaryServerInterceptor() grpc.UnaryServerInterceptor
	// MakeStreamServerInterceptor returns stream server interceptor.
	MakeStreamServerInterceptor() grpc.StreamServerInterceptor
}

type grpcImpl struct {
	backpressureOperator backpressure.Operator
	logger               logr.Logger
}

func (g *grpcImpl) MakeUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		allowed := g.backpressureOperator.AllowRequest()
		if allowed {
			return handler(ctx, req)
		}

		logger, err := logr.FromContext(ctx)
		if err != nil {
			logger = g.logger
		}

		logger.Info("request has been throttled")

		return nil, status.Error(codes.ResourceExhausted, "request has been throttled")
	}
}

func (g *grpcImpl) MakeStreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		allowed := g.backpressureOperator.AllowRequest()
		if allowed {
			return handler(srv, ss)
		}

		logger, err := logr.FromContext(ss.Context())
		if err != nil {
			logger = g.logger
		}

		logger.Info("request has been throttled")

		return status.Error(codes.ResourceExhausted, "request has been throttled")
	}
}
