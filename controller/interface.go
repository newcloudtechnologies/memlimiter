/*
 * Copyright (c) New Cloud Technologies, Ltd. 2013-2026.
 * Author: Vitaly Isaev <vitaly.isaev@myoffice.team>
 * License: https://github.com/newcloudtechnologies/memlimiter/blob/master/LICENSE
 */

package controller

import (
	"github.com/newcloudtechnologies/memlimiter/stats"
)

// Controller - generic memory consumption controller interface.
type Controller interface {
	GetStats() (*stats.ControllerStats, error)
	Quit()
}
