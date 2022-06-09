/*
 * Copyright (c) New Cloud Technologies, Ltd. 2013-2022.
 * Author: Vitaly Isaev <vitaly.isaev@myoffice.team>
 * License: https://github.com/newcloudtechnologies/memlimiter/blob/master/LICENSE
 */

package stats

import (
	"fmt"
)

// MemLimiterStats - top-level MemLimiter statistics data type.
type MemLimiterStats struct {
	// ControllerStats - memory budget controller statistics
	Controller *ControllerStats
	// Backpressure - backpressure subsystem statistics
	Backpressure *BackpressureStats
}

// ControllerStats - memory budget controller tracker.
type ControllerStats struct {
	// MemoryBudget - common memory budget information
	MemoryBudget *MemoryBudgetStats
	// NextGC - NextGC-aware controller statistics
	NextGC *ControllerNextGCStats
}

// MemoryBudgetStats - memory budget tracker.
type MemoryBudgetStats struct {
	// SpecialConsumers - specialized memory consumers (like CGO) statistics.
	SpecialConsumers *SpecialConsumersStats
	// RSSLimit - physical memory (RSS) consumption limit [bytes].
	RSSLimit uint64
	// GoAllocLimit - allocation limit for Go Runtime (with the except of CGO) [bytes].
	GoAllocLimit uint64
	// Utilization - memory budget utilization [percents]
	// (definition depends on a particular controller implementation).
	Utilization float64
}

// SpecialConsumersStats - specialized memory consumers statistics.
type SpecialConsumersStats struct {
	// Go - Go runtime managed consumers.
	Go map[string]uint64
	// Cgo - consumers residing beyond the Cgo border.
	Cgo map[string]uint64
}

// ControllerNextGCStats - NextGC-aware controller statistics.
type ControllerNextGCStats struct {
	// P - proportional component's output
	P float64
	// Output - final output
	Output float64
}

// BackpressureStats - backpressure subsystem statistics.
type BackpressureStats struct {
	// Throttling - throttling subsystem statistics.
	Throttling *ThrottlingStats
	// ControlParameters - control signal received from controller.
	ControlParameters *ControlParameters
}

// ThrottlingStats - throttling subsystem statistics.
type ThrottlingStats struct {
	// Passed - number of allowed requests.
	Passed uint64
	// Throttled - number of throttled requests.
	Throttled uint64
	// Total - total number of received requests (Passed + Throttled)
	Total uint64
}

// ControlParameters - вектор управляющих сигналов для системы.
type ControlParameters struct {
	// GOGC - value that will be used as a parameter for debug.SetGCPercent
	GOGC int
	// ThrottlingPercentage - percentage of requests that must be throttled on the middleware level (in range [0; 100])
	ThrottlingPercentage uint32
}

func (cp *ControlParameters) String() string {
	return fmt.Sprintf("gogc = %v, throttling_percentage = %v", cp.GOGC, cp.ThrottlingPercentage)
}

// ToKeysAndValues serializes struct for use in logr.Logger.
func (cp *ControlParameters) ToKeysAndValues() []interface{} {
	return []interface{}{
		"gogc", cp.GOGC,
		"throttling_percentage", cp.ThrottlingPercentage,
	}
}

// EqualsTo - comparator.
func (cp *ControlParameters) EqualsTo(other *ControlParameters) bool {
	return cp.GOGC == other.GOGC && cp.ThrottlingPercentage == other.ThrottlingPercentage
}
