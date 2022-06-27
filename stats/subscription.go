/*
 * Copyright (c) New Cloud Technologies, Ltd. 2013-2022.
 * Author: Vitaly Isaev <vitaly.isaev@myoffice.team>
 * License: https://github.com/newcloudtechnologies/memlimiter/blob/master/LICENSE
 */

package stats

import (
	"os"
	"runtime"
	"time"

	"github.com/go-logr/logr"
	"github.com/newcloudtechnologies/memlimiter/utils/breaker"
	"github.com/pkg/errors"
	"github.com/shirou/gopsutil/v3/process"
)

// ServiceStatsSubscription - service tracker subscription interface.
// There is a default implementation, but if you use Cgo in your application,
// it's strongly recommended to implement this interface on your own, because
// you need to provide custom tracker containing information on Cgo memory consumption.
type ServiceStatsSubscription interface {
	// Updates returns outgoing stream of service tracker.
	Updates() <-chan ServiceStats
	// Quit terminates program.
	Quit()
}

type subscriptionDefault struct {
	outChan chan ServiceStats
	breaker *breaker.Breaker
	logger  logr.Logger
	period  time.Duration
	pid     int32
}

func (s *subscriptionDefault) Updates() <-chan ServiceStats { return s.outChan }

func (s *subscriptionDefault) Quit() {
	s.breaker.ShutdownAndWait()
}

func (s *subscriptionDefault) makeServiceStats() (ServiceStats, error) {
	ms := &runtime.MemStats{}
	runtime.ReadMemStats(ms)

	pr, err := process.NewProcess(s.pid)
	if err != nil {
		return nil, errors.Wrap(err, "new pr")
	}

	processMemoryInfo, err := pr.MemoryInfoEx()
	if err != nil {
		return nil, errors.Wrap(err, "process memory info ex")
	}

	return serviceStatsDefault{
		rss:    processMemoryInfo.RSS,
		nextGC: ms.NextGC,
	}, nil
}

// NewSubscriptionDefault - default implementation of service tracker subscription.
func NewSubscriptionDefault(logger logr.Logger, period time.Duration) ServiceStatsSubscription {
	ss := &subscriptionDefault{
		outChan: make(chan ServiceStats),
		period:  period,
		breaker: breaker.NewBreakerWithInitValue(1),
		pid:     int32(os.Getpid()),
		logger:  logger,
	}

	go func() {
		ticker := time.NewTicker(period)
		defer ticker.Stop()

		defer ss.breaker.Dec()

		for {
			select {
			case <-ticker.C:
				out, err := ss.makeServiceStats()
				if err != nil {
					logger.Error(err, "make service stats")

					break
				}

				select {
				case ss.outChan <- out:
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
