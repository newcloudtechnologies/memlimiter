/*
 * Copyright (c) New Cloud Technologies, Ltd. 2013-2022.
 * Author: Vitaly Isaev <vitaly.isaev@myoffice.team>
 * License: https://github.com/newcloudtechnologies/memlimiter/blob/master/LICENSE
 */

package nextgc

import (
	"math"
	"testing"

	"github.com/go-logr/logr/testr"
	"github.com/stretchr/testify/require"
)

func TestComponentP(t *testing.T) {
	logger := testr.New(t)
	cmp := newComponentP(logger, &ComponentProportionalConfig{
		Coefficient: 1,
	})

	_, err := cmp.value(-1)
	require.Error(t, err)

	_, err = cmp.value(math.NaN())
	require.Error(t, err)

	_, err = cmp.value(2)
	require.NoError(t, err)

	out, err := cmp.value(math.Inf(1))
	require.NoError(t, err)
	require.InDelta(t, float64(100), out, 1e-9)

	_, err = cmp.value(math.Inf(-1))
	require.Error(t, err)
}
