/*
 * Copyright (c) New Cloud Technologies, Ltd. 2013-2022.
 * Author: Vitaly Isaev <vitaly.isaev@myoffice.team>
 * License: https://github.com/newcloudtechnologies/memlimiter/blob/master/LICENSE
 */

package backpressure

import (
	"testing"

	"github.com/go-logr/logr/testr"
	"github.com/newcloudtechnologies/memlimiter/stats"
	"github.com/stretchr/testify/require"
)

func TestOperator(t *testing.T) {
	logger := testr.New(t)
	notifications := make(chan *stats.MemLimiterStats, 1)

	op := NewOperator(logger, WithNotificationsOption(notifications))

	params := &stats.ControlParameters{
		GOGC:                 20,
		ThrottlingPercentage: 80,
	}

	err := op.SetControlParameters(params)
	require.NoError(t, err)

	notification := <-notifications

	require.Equal(t, params, notification.Backpressure.ControlParameters)
}
