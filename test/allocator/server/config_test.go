/*
 * Copyright (c) New Cloud Technologies, Ltd. 2013-2022.
 * Author: Vitaly Isaev <vitaly.isaev@myoffice.team>
 * License: https://github.com/newcloudtechnologies/memlimiter/blob/master/LICENSE
 */

package server

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfig(t *testing.T) {
	t.Run("empty endpoint", func(t *testing.T) {
		c := &Config{}
		require.Error(t, c.Prepare())
	})

	t.Run("empty tracker", func(t *testing.T) {
		c := &Config{
			Tracker:        nil,
			ListenEndpoint: "localhost:80",
		}
		require.Error(t, c.Prepare())
	})
}
