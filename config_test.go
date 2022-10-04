/*
 * Copyright (c) New Cloud Technologies, Ltd. 2013-2022.
 * Author: Vitaly Isaev <vitaly.isaev@myoffice.team>
 * License: https://github.com/newcloudtechnologies/memlimiter/blob/master/LICENSE
 */

package memlimiter

import (
	"testing"

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
}
