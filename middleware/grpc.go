/*
 * Copyright (c) New Cloud Technologies, Ltd. 2013-2026.
 * Author: Vitaly Isaev <vitaly.isaev@myoffice.team>
 * License: https://github.com/newcloudtechnologies/memlimiter/blob/master/LICENSE
 */

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
// at the time of GRPC server construction.
type GRPC interface {
	// MakeUnaryServerInterceptor returns unary server interceptor.
	MakeUnaryServerInterceptor() grpc.UnaryServerInterceptor
	// MakeStreamServerInterceptor returns stream server interceptor.
	MakeStreamServerInterceptor() grpc.StreamServerInterceptor
}

// grpcImpl is the implementation of the GRPC interface.
type grpcImpl struct {
	backpressureOperator backpressure.Operator
	logger               logr.Logger
}

// unknownGRPCMethod is a constant for the unknown GRPC method.
const unknownGRPCMethod = "<unknown>"

// MakeUnaryServerInterceptor returns a unary server interceptor.
func (g *grpcImpl) MakeUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		allowed := g.backpressureOperator.AllowRequest()
		if allowed {
			return handler(ctx, req)
		}

		logger, err := logr.FromContext(ctx)
		if err != nil {
			logger = g.logger
		}

		logger.Info("request has been throttled", "grpc_method", g.grpcMethodFromUnaryInfo(info))

		return nil, status.Error(codes.ResourceExhausted, "request has been throttled")
	}
}

// MakeStreamServerInterceptor returns a stream server interceptor.
func (g *grpcImpl) MakeStreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(
		srv any,
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

		logger.Info("request has been throttled", "grpc_method", g.grpcMethodFromStreamInfo(info))

		return status.Error(codes.ResourceExhausted, "request has been throttled")
	}
}

// grpcMethodFromUnaryInfo returns the GRPC method from the unary server info.
func (g *grpcImpl) grpcMethodFromUnaryInfo(info *grpc.UnaryServerInfo) string {
	if info == nil {
		return unknownGRPCMethod
	}

	method := info.FullMethod
	if method == "" {
		return unknownGRPCMethod
	}

	return method
}

// grpcMethodFromStreamInfo returns the GRPC method from the stream server info.
func (g *grpcImpl) grpcMethodFromStreamInfo(info *grpc.StreamServerInfo) string {
	if info == nil {
		return unknownGRPCMethod
	}

	method := info.FullMethod
	if method == "" {
		return unknownGRPCMethod
	}

	return method
}
