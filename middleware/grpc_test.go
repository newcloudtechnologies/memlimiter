/*
 * Copyright (c) New Cloud Technologies, Ltd. 2013-2022.
 * Author: Vitaly Isaev <vitaly.isaev@myoffice.team>
 * License: https://github.com/newcloudtechnologies/memlimiter/blob/master/LICENSE
 */

package middleware

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/newcloudtechnologies/memlimiter/stats"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type logRecord struct {
	msg string
	kv  []any
}

type captureSink struct {
	record *logRecord
	prefix []any
}

var _ logr.LogSink = (*captureSink)(nil)

func newCaptureSink() *captureSink {
	return &captureSink{
		record: &logRecord{},
	}
}

func (s *captureSink) Init(_ logr.RuntimeInfo) {}

func (s *captureSink) Enabled(_ int) bool { return true }

func (s *captureSink) Info(_ int, msg string, keysAndValues ...any) {
	all := append([]any{}, s.prefix...)
	all = append(all, keysAndValues...)

	s.record.msg = msg
	s.record.kv = all
}

func (s *captureSink) Error(_ error, _ string, _ ...any) {}

func (s *captureSink) WithName(_ string) logr.LogSink { return s }

func (s *captureSink) WithValues(keysAndValues ...any) logr.LogSink {
	prefix := append([]any{}, s.prefix...)
	prefix = append(prefix, keysAndValues...)

	return &captureSink{
		record: s.record,
		prefix: prefix,
	}
}

type backpressureOperatorStub struct {
	allow bool
}

func (b *backpressureOperatorStub) SetControlParameters(_ *stats.ControlParameters) error { return nil }

func (b *backpressureOperatorStub) AllowRequest() bool { return b.allow }

func (b *backpressureOperatorStub) GetStats() (*stats.BackpressureStats, error) {
	return &stats.BackpressureStats{}, nil
}

func (b *backpressureOperatorStub) Quit() {}

type serverStreamStub struct{}

func (s *serverStreamStub) SetHeader(_ metadata.MD) error { return nil }

func (s *serverStreamStub) SendHeader(_ metadata.MD) error { return nil }

func (s *serverStreamStub) SetTrailer(_ metadata.MD) {}

func (s *serverStreamStub) Context() context.Context { return context.Background() }

func (s *serverStreamStub) SendMsg(_ any) error { return nil }

func (s *serverStreamStub) RecvMsg(_ any) error { return nil }

func TestUnaryServerInterceptorLogsMethodOnThrottling(t *testing.T) {
	sink := newCaptureSink()
	g := &grpcImpl{
		backpressureOperator: &backpressureOperatorStub{allow: false},
		logger:               logr.New(sink),
	}

	interceptor := g.MakeUnaryServerInterceptor()

	handlerCalled := false
	_, err := interceptor(
		context.Background(),
		"struct{}{}",
		&grpc.UnaryServerInfo{FullMethod: "/test.Service/Unary"},
		func(_ context.Context, _ any) (any, error) {
			handlerCalled = true

			return "ok", nil
		},
	)

	require.False(t, handlerCalled)
	require.Error(t, err)

	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.ResourceExhausted, st.Code())

	method, ok := keyValueByName(sink.record.kv, "grpc_method")
	require.True(t, ok)
	require.Equal(t, "/test.Service/Unary", method)
}

func TestStreamServerInterceptorLogsMethodOnThrottling(t *testing.T) {
	sink := newCaptureSink()
	g := &grpcImpl{
		backpressureOperator: &backpressureOperatorStub{allow: false},
		logger:               logr.New(sink),
	}

	interceptor := g.MakeStreamServerInterceptor()

	handlerCalled := false
	err := interceptor(
		struct{}{},
		&serverStreamStub{},
		&grpc.StreamServerInfo{FullMethod: "/test.Service/Stream"},
		func(_ any, _ grpc.ServerStream) error {
			handlerCalled = true

			return nil
		},
	)

	require.False(t, handlerCalled)
	require.Error(t, err)

	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.ResourceExhausted, st.Code())

	method, ok := keyValueByName(sink.record.kv, "grpc_method")
	require.True(t, ok)
	require.Equal(t, "/test.Service/Stream", method)
}

func keyValueByName(kv []any, key string) (any, bool) {
	for i := 0; i+1 < len(kv); i += 2 {
		k, ok := kv[i].(string)
		if ok && k == key {
			return kv[i+1], true
		}
	}

	return nil, false
}
