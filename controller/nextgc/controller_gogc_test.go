/*
 * Copyright (c) New Cloud Technologies, Ltd. 2013-2026.
 * Author: Vitaly Isaev <vitaly.isaev@myoffice.team>
 * License: https://github.com/newcloudtechnologies/memlimiter/blob/master/LICENSE
 */

package nextgc

import (
	"testing"

	"github.com/newcloudtechnologies/memlimiter/backpressure"
	"github.com/newcloudtechnologies/memlimiter/stats"
	"github.com/stretchr/testify/require"
)

func TestUpdateControlParameterGOGC(t *testing.T) {
	t.Run("clamp to custom MinGOGC", func(t *testing.T) {
		c := &controllerImpl{
			sumValue:    99,
			utilization: 0.9,
			cfg: &ControllerConfig{
				DangerZoneGOGC: 50,
				MinGOGC:        10,
			},
			controlParameters: &stats.ControlParameters{},
		}

		c.updateControlParameterGOGC()

		require.Equal(t, 10, c.controlParameters.GOGC)
	})

	t.Run("clamp to default MinGOGC when MinGOGC is zero", func(t *testing.T) {
		c := &controllerImpl{
			sumValue:    99,
			utilization: 0.9,
			cfg: &ControllerConfig{
				DangerZoneGOGC: 50,
				MinGOGC:        0,
			},
			controlParameters: &stats.ControlParameters{},
		}

		c.updateControlParameterGOGC()

		require.Equal(t, defaultMinGOGC, c.controlParameters.GOGC)
	})

	t.Run("green zone keeps default GOGC", func(t *testing.T) {
		c := &controllerImpl{
			sumValue:    99,
			utilization: 0.1,
			cfg: &ControllerConfig{
				DangerZoneGOGC: 50,
				MinGOGC:        10,
			},
			controlParameters: &stats.ControlParameters{},
		}

		c.updateControlParameterGOGC()

		require.Equal(t, backpressure.DefaultGOGC, c.controlParameters.GOGC)
	})

	t.Run("value above MinGOGC is not clamped", func(t *testing.T) {
		c := &controllerImpl{
			sumValue:    22,
			utilization: 0.9,
			cfg: &ControllerConfig{
				DangerZoneGOGC: 50,
				MinGOGC:        10,
			},
			controlParameters: &stats.ControlParameters{},
		}

		c.updateControlParameterGOGC()

		require.Equal(t, 78, c.controlParameters.GOGC)
	})
}
