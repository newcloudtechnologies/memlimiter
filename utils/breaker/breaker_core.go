/*
 * Copyright (c) New Cloud Technologies, Ltd. 2013-2022.
 * Author: Vitaly Isaev <vitaly.isaev@myoffice.team>
 * License: https://github.com/newcloudtechnologies/memlimiter/blob/master/LICENSE
 */

package breaker

import (
	"runtime"
	"sync/atomic"
)

const (
	// operational accepts new tasks and allows Inc.
	operational int32 = iota + 1
	// shutdown rejects new tasks and allows Wait to drain in-flight work.
	shutdown
)

// breakerCore tracks active operations and shutdown state.
//
// It must not be copied after first use because it contains atomic fields.
type breakerCore struct {
	// count is the number of active tasks currently tracked by the breaker.
	count atomic.Int64
	// mode stores the current lifecycle state (operational or shutdown).
	mode atomic.Int32
}

// newBreakerCore creates a breaker core in operational mode.
func newBreakerCore() *breakerCore {
	b := &breakerCore{}
	b.mode.Store(operational)

	return b
}

// Inc increments the number of active tasks.
func (b *breakerCore) Inc() error {
	if !b.IsOperational() {
		return ErrShuttingDown
	}

	b.count.Add(1)

	return nil
}

// Dec decrements the number of active tasks.
func (b *breakerCore) Dec() {
	b.count.Add(-1)
}

// IsOperational reports whether the breaker is still accepting new tasks.
func (b *breakerCore) IsOperational() bool {
	return b.mode.Load() == operational
}

// Wait blocks until the number of active tasks reaches zero.
func (b *breakerCore) Wait() {
	if b.mode.Load() != shutdown {
		// Wait must be called only after Shutdown.
		// Panic matches Go sync primitives, which panic on API misuse.
		panic("cannot wait while breaker is operational, call Shutdown first")
	}

	for b.count.Load() != 0 {
		runtime.Gosched()
	}
}

// Shutdown moves the breaker to shutdown mode.
func (b *breakerCore) Shutdown() {
	_ = b.mode.CompareAndSwap(operational, shutdown)
}

// ShutdownAndWait moves the breaker to shutdown mode and waits for all tasks.
func (b *breakerCore) ShutdownAndWait() {
	b.Shutdown()
	b.Wait()
}

// Err mirrors context.Context semantics:
// nil while operational, otherwise ErrShutdown.
func (b *breakerCore) Err() error {
	if b.IsOperational() {
		return nil
	}

	return ErrShutdown
}
