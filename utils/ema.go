/*
 * Copyright (c) New Cloud Technologies, Ltd. 2013-2022.
 * Author: Vitaly Isaev <vitaly.isaev@myoffice.team>
 * License: https://github.com/newcloudtechnologies/memlimiter/blob/master/LICENSE
 */

package utils

import "sync"

// EMASmoother is a concurrency-safe exponential moving average calculator.
//
// It helps stabilize noisy measurements (for example, memory utilization)
// so control logic reacts to trend changes instead of short spikes.
type EMASmoother struct {
	// alpha is the smoothing coefficient.
	alpha float64
	// initialized is a flag indicating if the smoother has been initialized.
	initialized bool
	// value is the current smoothed value.
	value float64
	// mu is a mutex for synchronization.
	mu sync.Mutex
}

// NewEMASmoother creates a new exponential moving average calculator.
//
// The smoothing coefficient alpha is usually in (0; 1]:
// smaller alpha -> smoother but slower reaction,
// larger alpha  -> faster reaction but noisier output.
func NewEMASmoother(alpha float64) *EMASmoother {
	return &EMASmoother{alpha: alpha}
}

// Update adds a new sample and returns the current smoothed value.
//
// Formula:
//
//	S_t = alpha*X_t + (1-alpha)*S_{t-1}
func (e *EMASmoother) Update(value float64) float64 {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.initialized {
		e.value = value
		e.initialized = true

		return e.value
	}

	e.value = e.alpha*value + (1-e.alpha)*e.value

	return e.value
}
