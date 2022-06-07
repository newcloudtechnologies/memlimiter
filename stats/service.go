package stats

// ServiceStats represents the actual process statistics
type ServiceStats struct {
	// NextGC - current NextGC value [bytes]
	NextGC uint64
	// Custom - optional field, put here any service-specific information;
	// this object will be passed to utils.ConsumptionReporter.
	Custom interface{}
}
