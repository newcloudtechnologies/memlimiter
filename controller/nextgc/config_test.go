/*
 * Copyright (c) New Cloud Technologies, Ltd. 2013-2022.
 * Author: Vitaly Isaev <vitaly.isaev@myoffice.team>
 * License: https://github.com/newcloudtechnologies/memlimiter/blob/master/LICENSE
 */

package nextgc

import (
	"testing"

	"github.com/newcloudtechnologies/memlimiter/backpressure"
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

	t.Run("bad danger zone GOGC equal to 100", func(t *testing.T) {
		c := &ControllerConfig{
			RSSLimit:       bytes.Bytes{Value: 1},
			DangerZoneGOGC: 100,
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

	t.Run("bad danger zone throttling equal to 100", func(t *testing.T) {
		c := &ControllerConfig{
			RSSLimit:             bytes.Bytes{Value: 1},
			DangerZoneGOGC:       50,
			DangerZoneThrottling: 100,
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

	t.Run("default MinGOGC", func(t *testing.T) {
		c := &ControllerConfig{
			RSSLimit:             bytes.Bytes{Value: 1},
			DangerZoneGOGC:       50,
			DangerZoneThrottling: 90,
			Period:               duration.Duration{Duration: 1},
			ComponentProportional: &ComponentProportionalConfig{
				Coefficient: 1,
			},
		}

		require.NoError(t, c.Prepare())
		require.Equal(t, defaultMinGOGC, c.MinGOGC)
	})

	t.Run("invalid MinGOGC less than one", func(t *testing.T) {
		c := &ControllerConfig{
			RSSLimit:             bytes.Bytes{Value: 1},
			DangerZoneGOGC:       50,
			DangerZoneThrottling: 90,
			Period:               duration.Duration{Duration: 1},
			MinGOGC:              -1,
			ComponentProportional: &ComponentProportionalConfig{
				Coefficient: 1,
			},
		}

		require.Error(t, c.Prepare())
	})

	t.Run("invalid MinGOGC greater than default GOGC", func(t *testing.T) {
		c := &ControllerConfig{
			RSSLimit:             bytes.Bytes{Value: 1},
			DangerZoneGOGC:       50,
			DangerZoneThrottling: 90,
			Period:               duration.Duration{Duration: 1},
			MinGOGC:              backpressure.DefaultGOGC + 1,
			ComponentProportional: &ComponentProportionalConfig{
				Coefficient: 1,
			},
		}

		require.Error(t, c.Prepare())
	})

	t.Run("valid custom MinGOGC", func(t *testing.T) {
		const customMinGOGC = 25

		c := &ControllerConfig{
			RSSLimit:             bytes.Bytes{Value: 1},
			DangerZoneGOGC:       50,
			DangerZoneThrottling: 90,
			Period:               duration.Duration{Duration: 1},
			MinGOGC:              customMinGOGC,
			ComponentProportional: &ComponentProportionalConfig{
				Coefficient: 1,
			},
		}

		require.NoError(t, c.Prepare())
		require.Equal(t, customMinGOGC, c.MinGOGC)
	})

	t.Run("danger zone order is not strictly validated", func(t *testing.T) {
		c := &ControllerConfig{
			RSSLimit:             bytes.Bytes{Value: 1},
			DangerZoneGOGC:       90,
			DangerZoneThrottling: 50,
			Period:               duration.Duration{Duration: 1},
			MinGOGC:              10,
			ComponentProportional: &ComponentProportionalConfig{
				Coefficient: 1,
			},
		}

		require.NoError(t, c.Prepare())
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
