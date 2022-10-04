/*
 * Copyright (c) New Cloud Technologies, Ltd. 2013-2022.
 * Author: Vitaly Isaev <vitaly.isaev@myoffice.team>
 * License: https://github.com/newcloudtechnologies/memlimiter/blob/master/LICENSE
 */

package perf

import (
	"testing"

	"github.com/newcloudtechnologies/memlimiter/utils/config/bytes"
	"github.com/newcloudtechnologies/memlimiter/utils/config/duration"
	"github.com/stretchr/testify/require"
)

func TestConfig(t *testing.T) {
	t.Run("invalid endpoint", func(t *testing.T) {
		c := &Config{}
		require.Error(t, c.Prepare())
	})

	t.Run("invalid rps", func(t *testing.T) {
		c := &Config{
			Endpoint: "localhost:8080",
		}
		require.Error(t, c.Prepare())
	})

	t.Run("invalid load duration", func(t *testing.T) {
		c := &Config{
			Endpoint: "localhost:8080",
			RPS:      100,
		}
		require.Error(t, c.Prepare())
	})

	t.Run("invalid allocation size", func(t *testing.T) {
		c := &Config{
			Endpoint:     "localhost:8080",
			RPS:          100,
			LoadDuration: duration.Duration{Duration: 1},
		}
		require.Error(t, c.Prepare())
	})

	t.Run("invalid pause duration", func(t *testing.T) {
		c := &Config{
			Endpoint:       "localhost:8080",
			RPS:            100,
			LoadDuration:   duration.Duration{Duration: 1},
			AllocationSize: bytes.Bytes{Value: 100},
		}
		require.Error(t, c.Prepare())
	})

	t.Run("invalid request timeout duration", func(t *testing.T) {
		c := &Config{
			Endpoint:       "localhost:8080",
			RPS:            100,
			LoadDuration:   duration.Duration{Duration: 1},
			AllocationSize: bytes.Bytes{Value: 100},
			PauseDuration:  duration.Duration{Duration: 1},
		}
		require.Error(t, c.Prepare())
	})
}
