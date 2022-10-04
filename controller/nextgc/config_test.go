/*
 * Copyright (c) New Cloud Technologies, Ltd. 2013-2022.
 * Author: Vitaly Isaev <vitaly.isaev@myoffice.team>
 * License: https://github.com/newcloudtechnologies/memlimiter/blob/master/LICENSE
 */

package nextgc

import (
	"testing"

	"github.com/newcloudtechnologies/memlimiter/utils/config/bytes"
	"github.com/newcloudtechnologies/memlimiter/utils/config/duration"
	"github.com/stretchr/testify/require"
)

func TestComponentConfig(t *testing.T) {
	t.Run("bad RSS limit", func(t *testing.T) {
		c := &ControllerConfig{RSSLimit: bytes.Bytes{Value: 0}}
		require.Error(t, c.Prepare())
	})

	t.Run("bad danger zone GOGC", func(t *testing.T) {
		c := &ControllerConfig{
			RSSLimit:       bytes.Bytes{Value: 1},
			DangerZoneGOGC: 120,
		}
		require.Error(t, c.Prepare())
	})

	t.Run("bad danger zone throttling", func(t *testing.T) {
		c := &ControllerConfig{
			RSSLimit:             bytes.Bytes{Value: 1},
			DangerZoneGOGC:       50,
			DangerZoneThrottling: 120,
		}
		require.Error(t, c.Prepare())
	})

	t.Run("bad period", func(t *testing.T) {
		c := &ControllerConfig{
			RSSLimit:             bytes.Bytes{Value: 1},
			DangerZoneGOGC:       50,
			DangerZoneThrottling: 90,
			Period:               duration.Duration{Duration: 0},
		}
		require.Error(t, c.Prepare())
	})

	t.Run("bad component proportional", func(t *testing.T) {
		c := &ControllerConfig{
			RSSLimit:             bytes.Bytes{Value: 1},
			DangerZoneGOGC:       50,
			DangerZoneThrottling: 90,
			Period:               duration.Duration{Duration: 1},
		}
		require.Error(t, c.Prepare())
	})
}

func TestComponentProportionalConfig(t *testing.T) {
	t.Run("invalid proportional config", func(t *testing.T) {
		c := &ComponentProportionalConfig{
			Coefficient: 0,
		}
		require.Error(t, c.Prepare())
	})
}
