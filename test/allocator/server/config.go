package server

import (
	"github.com/newcloudtechnologies/memlimiter"
	"github.com/pkg/errors"
)

// Config - верхнеуровневая конфигурация сервиса Allocator.
type Config struct {
	MemLimiter *memlimiter.Config `json:"memlimiter"` //nolint:tagliatelle
	Server     *ServerConfig      `json:"server"`
}

// ServerConfig - конфигурация GRPC сервера.
type ServerConfig struct {
	ListenEndpoint string `json:"listen_endpoint"`
}

// Prepare - валидатор конфига.
func (c *Config) Prepare() error {
	if c.Server == nil {
		return errors.New("server is empty")
	}

	return nil
}
