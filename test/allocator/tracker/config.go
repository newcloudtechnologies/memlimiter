/*
 * Copyright (c) New Cloud Technologies, Ltd. 2013-2022.
 * Author: Vitaly Isaev <vitaly.isaev@myoffice.team>
 * License: https://github.com/newcloudtechnologies/memlimiter/blob/master/LICENSE
 */

package tracker

import (
	"github.com/newcloudtechnologies/memlimiter/utils/config/duration"
	"github.com/pkg/errors"
)

// Config is a configuration of a tracker.
type Config struct {
	Path   string            `json:"path"`
	Period duration.Duration `json:"period"`
}

// Prepare validates config.
func (c *Config) Prepare() error {
	if c.Path == "" {
		return errors.New("empty path")
	}

	if c.Period.Duration == 0 {
		return errors.New("empty period")
	}

	return nil
}
