package stats

import (
	"runtime"
	"time"

	"github.com/newcloudtechnologies/memlimiter/utils/breaker"
)

// ServiceStatsSubscription - service stats subscription interface.
// There is a default implementation, but if you use Cgo in your application,
// it's strongly recommended to implement this interface on your own, because
// you need to provide custom stats containing information on Cgo memory consumption.
type ServiceStatsSubscription interface {
	// Updates returns outgoing stream of service stats.
	Updates() <-chan ServiceStats
	// Quit terminates program.
	Quit()
}

type subscriptionDefault struct {
	outChan chan ServiceStats
	breaker *breaker.Breaker
	period  time.Duration
}

func (s *subscriptionDefault) Updates() <-chan ServiceStats { return s.outChan }

func (s *subscriptionDefault) Quit() {
	s.breaker.ShutdownAndWait()
}

func (s *subscriptionDefault) makeServiceStats() ServiceStats {
	ms := &runtime.MemStats{}
	runtime.ReadMemStats(ms)

	return serviceStatsDefault{nextGC: ms.NextGC}
}

// NewSubscriptionDefault - default implementation of service stats subscription.
func NewSubscriptionDefault(period time.Duration) ServiceStatsSubscription {
	ss := &subscriptionDefault{
		outChan: make(chan ServiceStats),
		period:  period,
		breaker: breaker.NewBreakerWithInitValue(1),
	}

	go func() {
		ticker := time.NewTicker(period)
		defer ticker.Stop()

		defer ss.breaker.Dec()

		for {
			select {
			case <-ticker.C:
				select {
				case ss.outChan <- ss.makeServiceStats():
				case <-ss.breaker.Done():
					return
				}
			case <-ss.breaker.Done():
				return
			}
		}
	}()

	return ss
}
