// Package memlimiter - memory budget control subsystem for Go services.
// It tracks memory budget utilization and tries to stabilize memory usage with
// backpressure (GC and request throttling) techniques.
package memlimiter
