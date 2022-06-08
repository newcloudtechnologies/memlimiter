package middleware

import (
	"github.com/go-logr/logr"
	"github.com/newcloudtechnologies/memlimiter/backpressure"
)

// Middleware - extendable type responsible for MemLimiter integration with
// various web and microservice frameworks.
type Middleware interface {
	GRPC() GRPC
	// TODO: add new frameworks here
}

type middlewareImpl struct {
	backpressureOperator backpressure.Operator
	logger               logr.Logger
}

func (m *middlewareImpl) GRPC() GRPC {
	return &grpcImpl{
		logger:               m.logger,
		backpressureOperator: m.backpressureOperator,
	}
}

// NewMiddleware creates new middleware instance.
func NewMiddleware(logger logr.Logger, operator backpressure.Operator) Middleware {
	return &middlewareImpl{
		logger:               logger,
		backpressureOperator: operator,
	}
}
