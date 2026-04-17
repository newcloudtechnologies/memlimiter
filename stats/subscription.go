/*
 * Copyright (c) New Cloud Technologies, Ltd. 2013-2022.
 * Author: Vitaly Isaev <vitaly.isaev@myoffice.team>
 * License: https://github.com/newcloudtechnologies/memlimiter/blob/master/LICENSE
 */

package stats

import (
	"fmt"
	"math"
	"os"
	"runtime"
	"time"

	"github.com/go-logr/logr"
	"github.com/newcloudtechnologies/memlimiter/utils/breaker"
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
}

func (s *subscriptionDefault) Updates() <-chan ServiceStats { return s.outChan }

func (s *subscriptionDefault) Quit() {
	s.breaker.ShutdownAndWait()
}

func (s *subscriptionDefault) makeServiceStats() (ServiceStats, error) {
	ms := &runtime.MemStats{}
	runtime.ReadMemStats(ms)

	pid, err := getCurrentPID()
	if err != nil {
		return nil, fmt.Errorf("get current pid: %w", err)
	}

	pr, err := process.NewProcess(pid)
	if err != nil {
		return nil, fmt.Errorf("new pr: %w", err)
	}

	processMemoryInfo, err := pr.MemoryInfoEx()
	if err != nil {
		return nil, fmt.Errorf("process memory info ex: %w", err)
	}

	return serviceStatsDefault{
		rss:    processMemoryInfo.RSS,
		nextGC: ms.NextGC,
	}, nil
}

func getCurrentPID() (int32, error) {
	pid := os.Getpid()
	if pid < 0 || pid > math.MaxInt32 {
		return 0, fmt.Errorf("pid is out of int32 range: %d", pid)
	}

	return int32(pid), nil
}

// NewSubscriptionDefault - default implementation of service tracker subscription.
func NewSubscriptionDefault(logger logr.Logger, period time.Duration) ServiceStatsSubscription {
	ss := &subscriptionDefault{
		outChan: make(chan ServiceStats),
		period:  period,
		breaker: breaker.NewBreakerWithInitValue(1),
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
