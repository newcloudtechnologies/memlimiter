/*
 * Copyright (c) New Cloud Technologies, Ltd. 2013-2022.
 * Author: Vitaly Isaev <vitaly.isaev@myoffice.team>
 * License: https://github.com/newcloudtechnologies/memlimiter/blob/master/LICENSE
 */

package tracker

import (
	"fmt"
)

type report struct {
	timestamp   string
	rss         uint64
	utilization float64
	gogc        int
	throttling  uint32
}

func (r *report) headers() []string {
	return []string{
		"timestamp",
		"rss",
		"utilization",
		"gogc",
		"throttling",
	}
}

func (r *report) toCsv() []string {
	return []string{
		r.timestamp,
		fmt.Sprint(r.rss),
		fmt.Sprint(r.utilization),
		fmt.Sprint(r.gogc),
		fmt.Sprint(r.throttling),
	}
}
