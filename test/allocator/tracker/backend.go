/*
 * Copyright (c) New Cloud Technologies, Ltd. 2013-2022.
 * Author: Vitaly Isaev <vitaly.isaev@myoffice.team>
 * License: https://github.com/newcloudtechnologies/memlimiter/blob/master/LICENSE
 */

package tracker

// backend is the interface that wraps the basic saveReport, getReports, and quit methods.
type backend interface {
	// saveReport saves a report to the backend.
	saveReport(report *Report) error
	// getReports gets the reports from the backend.
	getReports() ([]*Report, error)
	// quit quits the backend.
	quit()
}
