// Package nextgc provides the implementation of memory usage controller, which aims
// to keep Go Runtime NextGC value lower than the RSS consumption hard limit to prevent OOM errors.
package nextgc
