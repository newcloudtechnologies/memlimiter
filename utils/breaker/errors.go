/*
 * Copyright (c) New Cloud Technologies, Ltd. 2013-2026.
 * Author: Vitaly Isaev <vitaly.isaev@myoffice.team>
 * License: https://github.com/newcloudtechnologies/memlimiter/blob/master/LICENSE
 */

package breaker

import "errors"

var (
	// ErrShutdown indicates that the breaker has already shut down.
	ErrShutdown = errors.New("breaker is shut down")

	// ErrShuttingDown indicates that shutdown has started
	// and the breaker no longer accepts new tasks.
	ErrShuttingDown = errors.New("breaker is shutting down")
)
