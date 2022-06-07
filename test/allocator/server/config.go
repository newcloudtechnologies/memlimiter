package server

import (
	"github.com/newcloudtechnologies/memlimiter"
	"github.com/pkg/errors"
)

// Config - a top-level service configuration.
type Config struct {
	MemLimiter *memlimiter.Config `json:"memlimiter"` //nolint:tagliatelle
	Server     *ServerConfig      `json:"server"`
}

// ServerConfig - GRPC server configuration.
type ServerConfig struct {
	ListenEndpoint string `json:"listen_endpoint"`
}

// Prepare validates config.
func (c *Config) Prepare() error {
	if c.Server == nil {
		return errors.New("server is empty")
	}

	return nil
}
