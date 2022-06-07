package perf

import (
	"golang.org/x/time/rate"

	"github.com/newcloudtechnologies/memlimiter/utils/config/bytes"
	"github.com/newcloudtechnologies/memlimiter/utils/config/duration"
	"github.com/pkg/errors"
)

// Config - performance client configuration.
type Config struct {
	// Endpoint server address
	Endpoint string `json:"endpoint"`
	// RPS - target load [issued requests per second]
	RPS rate.Limit `json:"rps"`
	// LoadDuration - duration of the performance test session
	LoadDuration duration.Duration `json:"load_duration"`
	// AllocationSize - the size of allocations made on the server side during each request
	AllocationSize bytes.Bytes `json:"allocation_size"`
	// PauseDuration - duration of the pause in the request handler
	// on the server-side (to help allocations reside in server memory for a long time).
	PauseDuration duration.Duration `json:"pause_duration"`
	// RequestTimeout - server request timeout
	RequestTimeout duration.Duration `json:"request_timeout"`
}

// Prepare validates configuration.
func (c *Config) Prepare() error {
	if c.Endpoint == "" {
		return errors.New("empty endpoint")
	}

	if c.RPS == 0 {
		return errors.New("empty rps")
	}

	if c.LoadDuration.Duration == 0 {
		return errors.New("empty load duration")
	}

	if c.AllocationSize.Value == 0 {
		return errors.New("empty allocation size")
	}

	if c.PauseDuration.Duration == 0 {
		return errors.New("empty pause duration")
	}

	if c.RequestTimeout.Duration == 0 {
		return errors.New("empty request timeout")
	}

	return nil
}
