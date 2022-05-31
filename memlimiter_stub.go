package memlimiter

import (
	"context"

	"github.com/newcloudtechnologies/memlimiter/stats"
	"google.golang.org/grpc"
)

type memLimiterStub struct {
}

func (m memLimiterStub) Init(_ stats.Subscription) error { return nil }

func (m memLimiterStub) GetStats() (*stats.Memlimiter, error) {
	return nil, nil
}

func (m memLimiterStub) MakeUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		return handler(ctx, req)
	}
}

func (m memLimiterStub) MakeStreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		return handler(srv, ss)
	}
}

func (m memLimiterStub) Quit() {}
