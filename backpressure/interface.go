/*
 * Copyright (c) New Cloud Technologies, Ltd. 2013-2022.
 * Author: Vitaly Isaev <vitaly.isaev@myoffice.team>
 * License: https://github.com/newcloudtechnologies/memlimiter/blob/master/LICENSE
 */

package backpressure

import (
	"github.com/newcloudtechnologies/memlimiter/stats"
)

const (
	// DefaultGOGC - default GOGC value.
	DefaultGOGC = 100
	// NoThrottling - allow execution of all GRPC requests.
	NoThrottling = 0
	// FullThrottling - a complete ban on GRPC request execution.
	FullThrottling = 100
)

// Operator applies control signals to Go runtime and GRPC server.
type Operator interface {
	// SetControlParameters registers the actual value of control parameters.
	SetControlParameters(value *stats.ControlParameters) error
	// AllowRequest can be used by server middleware to check if it's possible to execute
	// a particular request.
	AllowRequest() bool
	// GetStats returns statistics of Backpressure subsystem.
	GetStats() (*stats.BackpressureStats, error)
}
