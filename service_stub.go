/*
 * Copyright (c) New Cloud Technologies, Ltd. 2013-2026.
 * Author: Vitaly Isaev <vitaly.isaev@myoffice.team>
 * License: https://github.com/newcloudtechnologies/memlimiter/blob/master/LICENSE
 */

package memlimiter

import (
	"sync/atomic"

	"github.com/newcloudtechnologies/memlimiter/middleware"
	"github.com/newcloudtechnologies/memlimiter/stats"
	"github.com/newcloudtechnologies/memlimiter/utils/breaker"
)

var _ Service = (*serviceStub)(nil)

// serviceStub doesn't perform active memory management, it just caches the latest statistics.
type serviceStub struct {
	latestStats       atomic.Value
	statsSubscription stats.ServiceStatsSubscription
	breaker           *breaker.Breaker
}

// newServiceStub constructs a new service stub.
func newServiceStub(statsSubscription stats.ServiceStatsSubscription) Service {
	if statsSubscription == nil {
		return &serviceStub{
			breaker: breaker.NewBreaker(),
		}
	}

	out := &serviceStub{
		statsSubscription: statsSubscription,
		breaker:           breaker.NewBreakerWithInitValue(1),
	}

	go out.loop()

	return out
}

// Middleware returns the middleware.
func (s *serviceStub) Middleware() middleware.Middleware {
	// TODO: return stub
	return nil
}

// Quit terminates the service stub gracefully.
func (s *serviceStub) Quit() {
	s.breaker.Shutdown()
	s.statsSubscription.Quit()
}

// GetStats returns the current stats.
func (s *serviceStub) GetStats() (*stats.MemLimiterStats, error) {
	if val := s.latestStats.Load(); val != nil {
		//nolint:forcetypeassert
		ss := val.(stats.ServiceStats)

		out := &stats.MemLimiterStats{
			Controller: &stats.ControllerStats{
				MemoryBudget: &stats.MemoryBudgetStats{
					RSSActual: ss.RSS(),
				},
			},
		}

		return out, nil
	}

	//nolint:nilnil // This is a stub.
	return nil, nil
}

// loop is the main loop of the service stub.
func (s *serviceStub) loop() {
	defer s.breaker.Dec()

	for {
		select {
		case record := <-s.statsSubscription.Updates():
			s.latestStats.Store(record)
		case <-s.breaker.Done():
			return
		}
	}
}
