/*
 * Copyright (c) New Cloud Technologies, Ltd. 2013-2026.
 * Author: Vitaly Isaev <vitaly.isaev@myoffice.team>
 * License: https://github.com/newcloudtechnologies/memlimiter/blob/master/LICENSE
 */

package breaker

import (
	"time"
)

// Breaker can be used to stop any subsystem with background tasks gracefully.
type Breaker struct {
	// breakerCore is the core of the breaker.
	*breakerCore

	// exitChan is the channel that is closed when the breaker is shut down.
	exitChan chan struct{}
}

// NewBreaker - default breaker constructor.
func NewBreaker() *Breaker {
	return &Breaker{
		breakerCore: newBreakerCore(),
		exitChan:    make(chan struct{}),
	}
}

// NewBreakerWithInitValue - alternative breaker constructor convenient for usage
// in pools and actors, when you know how many goroutines will work from the very beginning.
func NewBreakerWithInitValue(count int64) *Breaker {
	b := NewBreaker()
	b.count.Store(count)

	return b
}

// Shutdown switches breaker in shutdown mode.
func (b *Breaker) Shutdown() {
	if b.mode.CompareAndSwap(operational, shutdown) {
		// Notify channel subscribers about termination.
		close(b.exitChan)
	}
}

// ShutdownAndWait switches breakers in shutdown mode and
// waits for all background tasks to terminate.
func (b *Breaker) ShutdownAndWait() {
	b.Shutdown()
	b.Wait()
}

// Deadline is implemented for the sake of compatibility with context.Context.
func (b *Breaker) Deadline() (time.Time, bool) {
	return time.Time{}, false
}

// Value is implemented for the sake of compatibility with context.Context.
func (b *Breaker) Value(_ any) any {
	return nil
}

// Done returns channel which can be used in a manner similar to context.Context.Done().
func (b *Breaker) Done() <-chan struct{} {
	return b.exitChan
}
