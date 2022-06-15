/*
 * Copyright (c) New Cloud Technologies, Ltd. 2013-2022.
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

// serviceStub doesn't perform active memory management, it just caches the latest statistics
type serviceStub struct {
	latestStats       atomic.Value
	statsSubscription stats.ServiceStatsSubscription
	breaker           *breaker.Breaker
}

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

func (s *serviceStub) Middleware() middleware.Middleware {
	// FIXME: return stub
	return nil
}

func (s *serviceStub) GetStats() (*stats.MemLimiterStats, error) {
	if val := s.latestStats.Load(); val != nil {
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

	return nil, nil
}

func (s *serviceStub) Quit() {
	s.breaker.Shutdown()
	s.statsSubscription.Quit()
}

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
