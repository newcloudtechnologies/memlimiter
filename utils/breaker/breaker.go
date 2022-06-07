package breaker

import (
	"runtime"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
)

const (
	operational int32 = iota + 1
	shutdown
)

// Breaker can be used to stop any subsystem with background tasks gracefully.
type Breaker struct {
	count    int64
	mode     int32
	exitChan chan struct{}
}

// Inc increments number of tasks.
func (b *Breaker) Inc() error {
	if !b.isOperational() {
		return errors.New("shutdown in progress")
	}

	atomic.AddInt64(&b.count, 1)

	return nil
}

// Dec decrements number of tasks.
func (b *Breaker) Dec() {
	atomic.AddInt64(&b.count, -1)
}

// isOperational checks whether breaker is in operational mode.
func (b *Breaker) isOperational() bool { return atomic.LoadInt32(&b.mode) == operational }

// Wait blocks until the number of tasks becomes equal to zero.
func (b *Breaker) Wait() {
	if atomic.LoadInt32(&b.mode) != shutdown {
		panic("cannot wait on operational Breaker, turn it off first")
	}

	for {
		if atomic.LoadInt64(&b.count) == 0 {
			break
		}

		runtime.Gosched()
	}
}

// Shutdown switches breaker in shutdown mode.
func (b *Breaker) Shutdown() {
	if atomic.CompareAndSwapInt32(&b.mode, operational, shutdown) {
		// notify channel subscribers about termination
		close(b.exitChan)
	}
}

// ShutdownAndWait switches breakers in shutdown mode and
// waits for all background tasks to terminate.
func (b *Breaker) ShutdownAndWait() {
	b.Shutdown()
	b.Wait()
}

func (b *Breaker) Deadline() (deadline time.Time, ok bool) {
	return time.Time{}, false
}

func (b *Breaker) Value(key interface{}) interface{} { return nil }

// Done returns channel which can be used in a manner similar to context.Context.Done().
func (b *Breaker) Done() <-chan struct{} { return b.exitChan }

// Err returns error which can be used in a manner similar to context.Context.Done().
func (b *Breaker) Err() error {
	if b.isOperational() {
		return nil
	}

	return errors.New("breaker is not operational")
}

// NewBreaker - default breaker constructor.
func NewBreaker() *Breaker {
	return &Breaker{
		count:    0,
		mode:     operational,
		exitChan: make(chan struct{}),
	}
}

// NewBreakerWithInitValue - alternative breaker constructor convenient for usage
// in pools and actors, when you know how many goroutines will work from the very beginning.
func NewBreakerWithInitValue(value int64) *Breaker {
	b := NewBreaker()
	b.count = value

	return b
}
