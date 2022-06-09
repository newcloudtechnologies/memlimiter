/*
 * Copyright (c) New Cloud Technologies, Ltd. 2013-2022.
 * Author: Vitaly Isaev <vitaly.isaev@myoffice.team>
 * License: https://github.com/newcloudtechnologies/memlimiter/blob/master/LICENSE
 */

package server

import (
	"github.com/newcloudtechnologies/memlimiter"
	"github.com/newcloudtechnologies/memlimiter/test/allocator/tracker"
	"github.com/pkg/errors"
)

// Config - a top-level service configuration.
type Config struct {
	MemLimiter     *memlimiter.Config `json:"memLimiter"` //nolint:tagliatelle
	Tracker        *tracker.Config    `json:"tracker"`
	ListenEndpoint string             `json:"listen_endpoint"`
}

// Prepare validates config.
func (c *Config) Prepare() error {
	if c.ListenEndpoint == "" {
		return errors.New("listen endpoint is empty")
	}

	if c.Tracker == nil {
		return errors.New("empty tracker")
	}

	return nil
}
