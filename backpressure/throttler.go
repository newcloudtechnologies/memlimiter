package backpressure

import (
	"sync/atomic"

	"github.com/newcloudtechnologies/memlimiter/stats"
	"github.com/newcloudtechnologies/memlimiter/utils"
	"github.com/villenny/fastrand64-go"

	"github.com/pkg/errors"
)

type throttler struct {
	// group of request counters
	requestsTotal     utils.Counter
	requestsPassed    utils.Counter
	requestsThrottled utils.Counter

	// The following features of RNG are crucial for the backpressure subsystem:
	// 1. uniform distribution of the output;
	// 2. thread-safety;

	// In order to save time, we use a third party RNG library which is thread-safe,
	// however, there are some concerns on the distribution uniformity.
	// There are indicators that it's truly uniform because this RNG (providing only integer numbers)
	// was used in the uniformly distributed float RNG implementation: https://prng.di.unimi.it/

	// Here are the posts stating that (at least empirically) the distribution uniformity is
	// observed for all RNGs belonging to this family but one:
	// https://stackoverflow.com/questions/71050149/does-xoshiro-xoroshiro-prngs-provide-uniform-distribution
	// https://crypto.stackexchange.com/questions/98597
	rng *fastrand64.ThreadsafePoolRNG

	// число в диапазоне [0; 100], показывающее, какой процент запросов должен быть отбит
	threshold uint32
}

func (t *throttler) setThreshold(value uint32) error {
	if value > FullThrottling {
		return errors.New("implementation error: threshold value must belong to [0;100]")
	}

	atomic.StoreUint32(&t.threshold, value)

	return nil
}

func (t *throttler) getStats() *stats.ThrottlingStats {
	return &stats.ThrottlingStats{
		Total:     uint64(t.requestsTotal.Count()),
		Passed:    uint64(t.requestsPassed.Count()),
		Throttled: uint64(t.requestsThrottled.Count()),
	}
}

func (t *throttler) AllowRequest() bool {
	threshold := atomic.LoadUint32(&t.threshold)

	// if throttling is disabled, allow any request
	if threshold == 0 {
		t.requestsPassed.Inc(1)

		return true
	}

	// flip a coin in the range [0; 100]; if the actual value is less than the threshold value,
	// throttle the request, otherwise allow it.
	value := t.rng.Uint32n(FullThrottling)

	allowed := value >= threshold

	if allowed {
		t.requestsPassed.Inc(1)
	} else {
		t.requestsThrottled.Inc(1)
	}

	return allowed
}

func newThrottler() *throttler {
	requestsTotal := utils.NewCounter(nil)

	return &throttler{
		rng:               fastrand64.NewSyncPoolXoshiro256ssRNG(),
		requestsTotal:     requestsTotal,
		requestsPassed:    utils.NewCounter(requestsTotal),
		requestsThrottled: utils.NewCounter(requestsTotal),
	}
}
