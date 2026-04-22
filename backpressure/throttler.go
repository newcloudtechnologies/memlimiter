/*
 * Copyright (c) New Cloud Technologies, Ltd. 2013-2022.
 * Author: Vitaly Isaev <vitaly.isaev@myoffice.team>
 * License: https://github.com/newcloudtechnologies/memlimiter/blob/master/LICENSE
 */

package backpressure

import (
	"errors"
	"math/rand/v2"
	"sync/atomic"

	"github.com/newcloudtechnologies/memlimiter/stats"
	"github.com/newcloudtechnologies/memlimiter/utils"
)

// throttler is a struct that implements the throttler.
// It must not be copied after first use because it contains atomic fields.
// It is safe for concurrent use.
type throttler struct {
	// requestsTotal is the total number of requests.
	requestsTotal utils.Counter[uint64]
	// requestsPassed is the number of requests that were passed.
	requestsPassed utils.Counter[uint64]
	// requestsThrottled is the number of requests that were throttled.
	requestsThrottled utils.Counter[uint64]
	// threshold is the percentage of requests that should be throttled.
	// It must be in the range [0; 100].
	threshold atomic.Uint32
}

// newThrottler creates a new throttler.
func newThrottler() *throttler {
	requestsTotal := utils.NewUint64Counter(nil)

	return &throttler{
		requestsTotal:     requestsTotal,
		requestsPassed:    utils.NewUint64Counter(requestsTotal),
		requestsThrottled: utils.NewUint64Counter(requestsTotal),
	}
}

// AllowRequest checks if the request should be allowed.
func (t *throttler) AllowRequest() bool {
	threshold := t.threshold.Load()

	// If throttling is disabled, allow any request.
	if threshold == 0 {
		t.requestsPassed.Inc(1)

		return true
	}

	// Flip a coin in the range [0; 100].
	// If the actual value is less than the threshold value, throttle the request.
	// Otherwise, allow the request.
	// math/rand/v2 top-level functions are safe for concurrent use and provide
	// non-cryptographic uniformly distributed values, which is enough here.
	//nolint:gosec // Non-cryptographic RNG is intentional for probabilistic throttling decisions.
	value := rand.Uint32N(FullThrottling)

	allowed := value >= threshold

	if allowed {
		t.requestsPassed.Inc(1)
	} else {
		t.requestsThrottled.Inc(1)
	}

	return allowed
}

// setThreshold sets the threshold for the throttler.
func (t *throttler) setThreshold(value uint32) error {
	if value > FullThrottling {
		return errors.New("implementation error: threshold value must belong to [0;100]")
	}

	t.threshold.Store(value)

	return nil
}

// getStats returns the statistics of the throttler.
func (t *throttler) getStats() *stats.ThrottlingStats {
	return &stats.ThrottlingStats{
		Total:     t.requestsTotal.Count(),
		Passed:    t.requestsPassed.Count(),
		Throttled: t.requestsThrottled.Count(),
	}
}
