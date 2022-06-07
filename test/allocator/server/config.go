package server

import (
	"github.com/newcloudtechnologies/memlimiter"
	"github.com/pkg/errors"
)

// Config - a top-level service configuration.
type Config struct {
	MemLimiter     *memlimiter.Config `json:"memlimiter"` //nolint:tagliatelle
	ListenEndpoint string             `json:"listen_endpoint"`
}

// Prepare validates config.
func (c *Config) Prepare() error {
	if c.ListenEndpoint == "" {
		return errors.New("listen endpoint is empty")
	}

	return nil
}
