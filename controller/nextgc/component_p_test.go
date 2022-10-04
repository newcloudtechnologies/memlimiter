/*
 * Copyright (c) New Cloud Technologies, Ltd. 2013-2022.
 * Author: Vitaly Isaev <vitaly.isaev@myoffice.team>
 * License: https://github.com/newcloudtechnologies/memlimiter/blob/master/LICENSE
 */

package nextgc

import (
	"testing"

	"github.com/go-logr/logr/testr"
	"github.com/stretchr/testify/require"
)

func TestComponentP(t *testing.T) {
	logger := testr.New(t)
	cmp := &componentP{logger: logger}

	_, err := cmp.value(-1)
	require.Error(t, err)

	_, err = cmp.value(2)
	require.NoError(t, err)
}
