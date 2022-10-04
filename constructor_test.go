/*
 * Copyright (c) New Cloud Technologies, Ltd. 2013-2022.
 * Author: Vitaly Isaev <vitaly.isaev@myoffice.team>
 * License: https://github.com/newcloudtechnologies/memlimiter/blob/master/LICENSE
 */

package memlimiter

import (
	"testing"
	"time"

	"github.com/go-logr/logr/testr"
	"github.com/newcloudtechnologies/memlimiter/stats"
	"github.com/stretchr/testify/require"
)

func TestConstructor(t *testing.T) {
	t.Run("stub", func(t *testing.T) {
		logger := testr.New(t)

		const delay = 10 * time.Millisecond

		subscription := stats.NewSubscriptionDefault(logger, delay)
		defer subscription.Quit()

		service, err := NewServiceFromConfig(
			logger,
			nil, // use stub instead of real service
			WithServiceStatsSubscription(subscription),
		)
		require.NoError(t, err)

		defer service.Quit()

		ss, err := service.GetStats()
		require.NoError(t, err)
		require.Nil(t, ss)

		time.Sleep(2 * delay)

		ss, err = service.GetStats()
		require.NoError(t, err)
		require.NotNil(t, ss)

		require.Nil(t, service.Middleware())
	})
}
