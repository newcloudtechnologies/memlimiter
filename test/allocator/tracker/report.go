/*
 * Copyright (c) New Cloud Technologies, Ltd. 2013-2022.
 * Author: Vitaly Isaev <vitaly.isaev@myoffice.team>
 * License: https://github.com/newcloudtechnologies/memlimiter/blob/master/LICENSE
 */

package tracker

import (
	"fmt"
	"strconv"
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
		strconv.FormatUint(r.RSS, 10),
		fmt.Sprint(r.Utilization),
		strconv.Itoa(r.GOGC),
		strconv.FormatUint(uint64(r.Throttling), 10),
	}
}
