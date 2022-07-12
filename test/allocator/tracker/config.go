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
	BackendFile   *ConfigBackendFile   `json:"backend_file"`
	BackendMemory *ConfigBackendMemory `json:"backend_memory"`
	Period        duration.Duration    `json:"period"`
}

type ConfigBackendFile struct {
	Path string `json:"path"`
}

func (c *ConfigBackendFile) Prepare() error {
	if c.Path == "" {
		return errors.New("empty path")
	}

	return nil
}

type ConfigBackendMemory struct {
}

// Prepare validates config.
func (c *Config) Prepare() error {
	if c.BackendFile == nil && c.BackendMemory == nil {
		return errors.New("empty backend sections")
	}

	if c.BackendFile != nil && c.BackendMemory != nil {
		return errors.New("more than one non-empty backend section")
	}

	if c.Period.Duration == 0 {
		return errors.New("empty period")
	}

	return nil
}
