/*
 * Copyright (c) New Cloud Technologies, Ltd. 2013-2026.
 * Author: Vitaly Isaev <vitaly.isaev@myoffice.team>
 * License: https://github.com/newcloudtechnologies/memlimiter/blob/master/LICENSE
 */

package memlimiter

import (
	"errors"
	"math"

	"github.com/newcloudtechnologies/memlimiter/controller/nextgc"
	"github.com/newcloudtechnologies/memlimiter/utils/config/bytes"
)

// Config - high-level MemLimiter config.
type Config struct {
	// GoMemoryLimit optionally sets Go runtime soft memory limit via debug.SetMemoryLimit.
	// Zero means disabled.
	GoMemoryLimit bytes.Bytes `json:"go_memory_limit"`
	// ControllerNextGC - NextGC-based controller
	ControllerNextGC *nextgc.ControllerConfig `json:"controller_nextgc"` //nolint:tagliatelle
	// TODO:
	//  if new controller implementation appears, put its config here and make switch in Prepare()
	//  (only one subsection must be not nil).
}

// Prepare validates config.
func (c *Config) Prepare() error {
	if c == nil {
		// This means that user wants to use stub instead of real memlimiter
		return nil
	}

	if c.ControllerNextGC == nil {
		return errors.New("empty ControllerNextGC")
	}

	if c.GoMemoryLimit.Value > uint64(math.MaxInt64) {
		return errors.New("GoMemoryLimit exceeds int64 range")
	}

	return nil
}
