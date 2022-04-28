/*
 * Copyright (c) New Cloud Technologies, Ltd. 2013-2022.
 * Author: Vitaly Isaev <vitaly.isaev@myoffice.team>
 * License: https://github.com/newcloudtechnologies/memlimiter/blob/master/LICENSE
 */

// Package memlimiter - memory budget control subsystem for Go services.
// It tracks memory budget utilization and tries to stabilize memory usage with
// backpressure (GC and request throttling) techniques.
package memlimiter
