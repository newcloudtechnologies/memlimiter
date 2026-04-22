/*
 * Copyright (c) New Cloud Technologies, Ltd. 2013-2026.
 * Author: Vitaly Isaev <vitaly.isaev@myoffice.team>
 * License: https://github.com/newcloudtechnologies/memlimiter/blob/master/LICENSE
 */

package memlimiter

import (
	"math"
	"testing"

	"github.com/newcloudtechnologies/memlimiter/controller/nextgc"
	"github.com/newcloudtechnologies/memlimiter/utils/config/bytes"
	"github.com/stretchr/testify/require"
)

func TestConfig(t *testing.T) {
	t.Run("config nil", func(t *testing.T) {
		var c *Config
		require.NoError(t, c.Prepare())
	})

	t.Run("controller config nil", func(t *testing.T) {
		c := &Config{ControllerNextGC: nil}
		require.Error(t, c.Prepare())
	})

	t.Run("go memory limit in range", func(t *testing.T) {
		c := &Config{
			ControllerNextGC: &nextgc.ControllerConfig{},
			GoMemoryLimit:    bytes.Bytes{Value: uint64(math.MaxInt64)},
		}
		require.NoError(t, c.Prepare())
	})

	t.Run("go memory limit out of range", func(t *testing.T) {
		c := &Config{
			ControllerNextGC: &nextgc.ControllerConfig{},
			GoMemoryLimit:    bytes.Bytes{Value: uint64(math.MaxInt64) + 1},
		}
		require.Error(t, c.Prepare())
	})
}
