/*
 * Copyright (c) New Cloud Technologies, Ltd. 2013-2022.
 * Author: Vitaly Isaev <vitaly.isaev@myoffice.team>
 * License: https://github.com/newcloudtechnologies/memlimiter/blob/master/LICENSE
 */

package memlimiter

import (
	"github.com/newcloudtechnologies/memlimiter/middleware"
	"github.com/newcloudtechnologies/memlimiter/stats"
)

// Service - a high-level interface for a memory usage control subsystem.
type Service interface {
	Middleware() middleware.Middleware
	GetStats() (*stats.MemLimiterStats, error)
	// Quit terminates service gracefully.
	Quit()
}
