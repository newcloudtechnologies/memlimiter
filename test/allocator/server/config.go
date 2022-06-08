/*
 * Copyright (c) New Cloud Technologies, Ltd. 2013-2022.
 * Author: Vitaly Isaev <vitaly.isaev@myoffice.team>
 * License: https://github.com/newcloudtechnologies/memlimiter/blob/master/LICENSE
 */

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
