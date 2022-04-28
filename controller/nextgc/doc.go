/*
 * Copyright (c) New Cloud Technologies, Ltd. 2013-2022.
 * Author: Vitaly Isaev <vitaly.isaev@myoffice.team>
 * License: https://github.com/newcloudtechnologies/memlimiter/blob/master/LICENSE
 */

// Package nextgc provides the implementation of memory usage controller, which aims
// to keep Go Runtime NextGC value lower than the RSS consumption hard limit to prevent OOM errors.
package nextgc
