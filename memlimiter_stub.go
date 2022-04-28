package memlimiter

import (
	"context"

	"google.golang.org/grpc"

	"gitlab.stageoffice.ru/UCS-COMMON/schemagen-go/v41/servus/stats/v1"

	"gitlab.stageoffice.ru/UCS-PLATFORM/servus"
	"gitlab.stageoffice.ru/UCS-PLATFORM/servus/feeder"
)

type memLimiterStub struct {
}

func (m memLimiterStub) Init(_ servus.Servus) error { return nil }

type stubFeeder struct{}

func (s stubFeeder) String() string {
	return "memlimiter_stub"
}

func (s stubFeeder) Feed(_ context.Context, _ *stats.ServiceStats) error {
	return nil
}

func (m memLimiterStub) ServusFeeder() feeder.Feeder { return stubFeeder{} }

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
