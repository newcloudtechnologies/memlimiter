/*
 * Copyright (c) New Cloud Technologies, Ltd. 2013-2022.
 * Author: Vitaly Isaev <vitaly.isaev@myoffice.team>
 * License: https://github.com/newcloudtechnologies/memlimiter/blob/master/LICENSE
 */

package stats

// ServiceStats represents the actual process statistics.
type ServiceStats interface {
	// RSS returns current RSS value [bytes]
	RSS() uint64
	// NextGC returns current NextGC value [bytes]
	NextGC() uint64
	// ConsumptionReport provides statistical information about the predefined memory consumers that contribute
	// significant part in process' overall memory consumption (caches, memory pools and other large structures).
	// It's mandatory to fill this report if you have large caches on Go side or if you allocate a lot beyond Cgo borders.
	// But if your service is simple, feel free to return nil.
	ConsumptionReport() *ConsumptionReport
}

// ConsumptionReport - report on memory consumption contributed by predefined data structures living during the
// whole application life-time (caches, memory pools and other large structures).
type ConsumptionReport struct {
	// Go - memory consumption contributed by structures managed by Go allocator.
	// [key - arbitrary string, value - bytes]
	Go map[string]uint64
	// Cgo - memory consumption contributed by structures managed by Cgo allocator.
	// [key - arbitrary string, value - bytes]
	Cgo map[string]uint64
}

var _ ServiceStats = (*serviceStatsDefault)(nil)

type serviceStatsDefault struct {
	rss    uint64
	nextGC uint64
}

func (s serviceStatsDefault) RSS() uint64 { return s.rss }

func (s serviceStatsDefault) NextGC() uint64 { return s.nextGC }

func (s serviceStatsDefault) ConsumptionReport() *ConsumptionReport {
	// don't forget to put real report of your service's memory consumption in your own implementation
	return nil
}
