package backpressure

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"gitlab.stageoffice.ru/UCS-PLATFORM/servus/stats/metrics"
)

func TestThrottler(t *testing.T) {
	// в каждом из тестов запускаем параллельно 1000 запросов, какая-то часть из них должна быть подавлена,
	// а какая-то - пройти
	const (
		requests = 1000
	)

	// С шагом в 10%
	for i := 0; i < 10; i++ {
		throttlingLevel := uint32(i) * 10

		t.Run(fmt.Sprintf("throttling level = %v", throttlingLevel), func(t *testing.T) {
			th := newThrottler()

			err := th.setThreshold(throttlingLevel)
			require.NoError(t, err)

			wg := &sync.WaitGroup{}
			wg.Add(requests)

			failed := metrics.NewCounter(nil)
			succeeded := metrics.NewCounter(nil)

			for i := 0; i < requests; i++ {
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
			require.Equal(t, int64(requests), total)

			failedShareExpected := float64(throttlingLevel) / float64(100)
			failedShareActual := float64(failed.Count()) / float64(total)
			succeededShareExpected := 1 - failedShareExpected
			succeededShareActual := float64(succeeded.Count()) / float64(total)

			// либо длина выборки недостаточная, либо распределение ГСЧ не совсем равномерное (см. комментарии к имплементации класса)
			// но длину выборки увеличивать нехорошо - ведь это не нагрузочные тесты, поэтому просто вводим погрешность
			require.InDelta(
				t,
				failedShareExpected,
				failedShareActual,
				0.055,
				fmt.Sprintf("expected = %v, actual = %v", failedShareExpected, failedShareActual),
			)
			require.InDelta(
				t,
				succeededShareExpected,
				succeededShareActual,
				0.055,
				fmt.Sprintf("expected = %v, actual = %v", succeededShareExpected, succeededShareActual),
			)

			// проверка внутренних счётчиков
			require.Equal(t, total, th.requestsTotal.Count())
			require.Equal(t, failed.Count(), th.requestsThrottled.Count())
			require.Equal(t, succeeded.Count(), th.requestsPassed.Count())
		})
	}
}

/*
$ go test -bench=. -benchtime=10s ./utils/memlimiter/backpressure
goos: linux
goarch: amd64
pkg: gitlab.stageoffice.ru/UCS-PLATFORM/dispersed-object-store/v4/utils/memlimiter/backpressure
cpu: 11th Gen Intel(R) Core(TM) i7-1165G7 @ 2.80GHz
BenchmarkThrottler/throttling_level_=_0-8         	   52140	    227815 ns/op
BenchmarkThrottler/throttling_level_=_50-8        	   52014	    228290 ns/op
BenchmarkThrottler/throttling_level_=_100-8       	   51867	    284538 ns/op.
*/
func BenchmarkThrottler(b *testing.B) {
	const requests = 1000

	for _, throttlingLevel := range []uint32{0, 50, 100} {
		throttlingLevel := throttlingLevel

		b.Run(fmt.Sprintf("throttling level = %v", throttlingLevel), func(b *testing.B) {
			th := newThrottler()

			err := th.setThreshold(throttlingLevel)
			if err != nil {
				b.Fatal(err)
			}

			var allowed bool

			for k := 0; k < b.N; k++ {
				wg := &sync.WaitGroup{}

				b.StartTimer()

				for i := 0; i < requests; i++ {
					wg.Add(1)

					go func() {
						defer wg.Done()
						allowed = th.AllowRequest() // присваиваем значение, чтобы компилятор не выпилил вызов
					}()
				}
				wg.Wait()

				b.StopTimer()
			}

			_ = allowed
		})
	}
}
