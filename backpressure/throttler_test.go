/*
 * Copyright (c) New Cloud Technologies, Ltd. 2013-2026.
 * Author: Vitaly Isaev <vitaly.isaev@myoffice.team>
 * License: https://github.com/newcloudtechnologies/memlimiter/blob/master/LICENSE
 */

package backpressure

import (
	"fmt"
	"sync"
	"testing"

	"github.com/newcloudtechnologies/memlimiter/utils"
	"github.com/stretchr/testify/require"
)

func TestThrottler(t *testing.T) {
	// Launch 1000 requests concurrently in each test. Some of them must be throttled, others should be allowed.
	const (
		requests = 1000
	)

	// Check different throttling levels with 10% step.
	for i := range 10 {
		throttlingLevel := uint32(i) * 10

		t.Run(fmt.Sprintf("throttling level = %v", throttlingLevel), func(t *testing.T) {
			th := newThrottler()

			err := th.setThreshold(throttlingLevel)
			require.NoError(t, err)

			wg := &sync.WaitGroup{}
			wg.Add(requests)

			failed := utils.NewUint64Counter(nil)
			succeeded := utils.NewUint64Counter(nil)

			for range requests {
				go func() {
					defer wg.Done()

					ok := th.AllowRequest()
					if ok {
						succeeded.Inc(1)
					} else {
						failed.Inc(1)
					}
				}()
			}

			wg.Wait()

			total := failed.Count() + succeeded.Count()
			require.Equal(t, uint64(requests), total)

			failedShareExpected := float64(throttlingLevel) / float64(100)
			failedShareActual := float64(failed.Count()) / float64(total)
			succeededShareExpected := 1 - failedShareExpected
			succeededShareActual := float64(succeeded.Count()) / float64(total)

			// Either sample length is not sufficient, or RNG distribution is not exactly uniform
			// (look through the comments in throttler.go).
			// We cannot increase sample length, because this is unit rather than performance tests,
			// so we introduce small inaccuracy.
			require.InDelta(
				t,
				failedShareExpected,
				failedShareActual,
				0.055,
				"expected = %v, actual = %v", failedShareExpected, failedShareActual,
			)
			require.InDelta(
				t,
				succeededShareExpected,
				succeededShareActual,
				0.055,
				"expected = %v, actual = %v", succeededShareExpected, succeededShareActual,
			)

			// Check internal counters.
			require.Equal(t, total, th.requestsTotal.Count())
			require.Equal(t, failed.Count(), th.requestsThrottled.Count())
			require.Equal(t, succeeded.Count(), th.requestsPassed.Count())
		})
	}
}

/*
go test -bench=. -benchtime=10s ./backpressure
goos: linux
goarch: amd64              consumption_reporter.go    doc.go                     mock.go
pkg: github.com/newcloudtechnologies/memlimiter/backpressure
cpu: AMD Ryzen 7 2700X Eight-Core Processor
BenchmarkThrottler/throttling_level_=_0-16                 22977            542772 ns/op
BenchmarkThrottler/throttling_level_=_50-16                22722            508701 ns/op
BenchmarkThrottler/throttling_level_=_100-16               22220            488162 ns/op
PASS
ok      github.com/newcloudtechnologies/memlimiter/backpressure 57.747s.
*/
func BenchmarkThrottler(b *testing.B) {
	const requests = 1000

	for _, throttlingLevel := range []uint32{0, 50, 100} {
		b.Run(fmt.Sprintf("throttling level = %v", throttlingLevel), func(b *testing.B) {
			th := newThrottler()

			err := th.setThreshold(throttlingLevel)
			if err != nil {
				b.Fatal(err)
			}

			var allowed bool

			for range b.N {
				wg := &sync.WaitGroup{}

				b.StartTimer()

				for range requests {
					wg.Go(func() {
						// assign result to fictive variable to disallow compiler to optimize out function call
						allowed = th.AllowRequest()
					})
				}

				wg.Wait()

				b.StopTimer()
			}

			_ = allowed
		})
	}
}
