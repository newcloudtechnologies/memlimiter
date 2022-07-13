/*
 * Copyright (c) New Cloud Technologies, Ltd. 2013-2022.
 * Author: Vitaly Isaev <vitaly.isaev@myoffice.team>
 * License: https://github.com/newcloudtechnologies/memlimiter/blob/master/LICENSE
 */

package tracker

type backendMemory struct {
	reports []*Report
}

func (b *backendMemory) saveReport(r *Report) error {
	b.reports = append(b.reports, r)

	return nil
}

func (b *backendMemory) getReports() ([]*Report, error) {
	out := make([]*Report, len(b.reports))
	copy(out, b.reports)

	return out, nil
}

func (b *backendMemory) quit() {}

func newBackendMemory() backend {
	return &backendMemory{}
}
