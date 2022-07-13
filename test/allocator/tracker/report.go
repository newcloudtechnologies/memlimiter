/*
 * Copyright (c) New Cloud Technologies, Ltd. 2013-2022.
 * Author: Vitaly Isaev <vitaly.isaev@myoffice.team>
 * License: https://github.com/newcloudtechnologies/memlimiter/blob/master/LICENSE
 */

package tracker

import (
	"fmt"
)

// Report is a memory consumption report (used only for tests).
type Report struct {
	Timestamp   string
	RSS         uint64
	Utilization float64
	GOGC        int
	Throttling  uint32
}

func (r *Report) headers() []string {
	return []string{
		"timestamp",
		"rss",
		"utilization",
		"gogc",
		"throttling",
	}
}

func (r *Report) toCsv() []string {
	return []string{
		r.Timestamp,
		fmt.Sprint(r.RSS),
		fmt.Sprint(r.Utilization),
		fmt.Sprint(r.GOGC),
		fmt.Sprint(r.Throttling),
	}
}
