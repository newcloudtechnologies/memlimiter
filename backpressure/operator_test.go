/*
 * Copyright (c) New Cloud Technologies, Ltd. 2013-2026.
 * Author: Vitaly Isaev <vitaly.isaev@myoffice.team>
 * License: https://github.com/newcloudtechnologies/memlimiter/blob/master/LICENSE
 */

package backpressure

import (
	"runtime/debug"
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

func TestOperatorQuitRestoresGOGC(t *testing.T) {
	const expectedInitialGOGC = 73

	originalBeforeTest := debug.SetGCPercent(expectedInitialGOGC)
	defer debug.SetGCPercent(originalBeforeTest)

	logger := testr.New(t)
	op := NewOperator(logger)

	err := op.SetControlParameters(&stats.ControlParameters{
		GOGC:                 21,
		ThrottlingPercentage: NoThrottling,
	})
	require.NoError(t, err)

	op.Quit()

	prev := debug.SetGCPercent(expectedInitialGOGC)
	require.Equal(t, expectedInitialGOGC, prev)
}
