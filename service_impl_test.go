/*
 * Copyright (c) New Cloud Technologies, Ltd. 2013-2022.
 * Author: Vitaly Isaev <vitaly.isaev@myoffice.team>
 * License: https://github.com/newcloudtechnologies/memlimiter/blob/master/LICENSE
 */

package memlimiter

import (
	"runtime/debug"
	"testing"
	"time"

	"github.com/go-logr/logr/testr"
	"github.com/newcloudtechnologies/memlimiter/backpressure"
	"github.com/newcloudtechnologies/memlimiter/controller"
	"github.com/newcloudtechnologies/memlimiter/controller/nextgc"
	"github.com/newcloudtechnologies/memlimiter/stats"
	"github.com/newcloudtechnologies/memlimiter/utils/config/bytes"
	"github.com/newcloudtechnologies/memlimiter/utils/config/duration"
	"github.com/stretchr/testify/require"
)

var (
	_ controller.Controller          = (*controllerStub)(nil)
	_ stats.ServiceStatsSubscription = (*serviceStatsSubscriptionStub)(nil)
	_ backpressure.Operator          = (*backpressureOperatorStub)(nil)
)

type controllerStub struct {
	quitCalled bool
}

type serviceStatsSubscriptionStub struct {
	quitCalled bool
}

type backpressureOperatorStub struct {
	quitCalled bool
}

func (c *controllerStub) GetStats() (*stats.ControllerStats, error) {
	return &stats.ControllerStats{}, nil
}

func (c *controllerStub) Quit() { c.quitCalled = true }

func (s *serviceStatsSubscriptionStub) Updates() <-chan stats.ServiceStats { return nil }

func (s *serviceStatsSubscriptionStub) Quit() { s.quitCalled = true }

func (b *backpressureOperatorStub) SetControlParameters(_ *stats.ControlParameters) error { return nil }

func (b *backpressureOperatorStub) AllowRequest() bool { return true }

func (b *backpressureOperatorStub) GetStats() (*stats.BackpressureStats, error) {
	return &stats.BackpressureStats{}, nil
}

func (b *backpressureOperatorStub) Quit() { b.quitCalled = true }

func TestServiceImplQuit(t *testing.T) {
	logger := testr.New(t)

	c := &controllerStub{}
	ss := &serviceStatsSubscriptionStub{}
	bp := &backpressureOperatorStub{}

	s := &serviceImpl{
		controller:           c,
		statsSubscription:    ss,
		backpressureOperator: bp,
		logger:               logger,
	}

	s.Quit()

	require.True(t, c.quitCalled)
	require.True(t, ss.quitCalled)
	require.True(t, bp.quitCalled)
}

func TestNewServiceImplGoMemoryLimitLifecycle(t *testing.T) {
	logger := testr.New(t)

	const (
		initialLimit  int64  = 512 << 20
		configuredMem uint64 = 256 << 20
	)

	previousBeforeTest := debug.SetMemoryLimit(initialLimit)
	defer debug.SetMemoryLimit(previousBeforeTest)

	require.Equal(t, initialLimit, debug.SetMemoryLimit(-1))

	cfg := &Config{
		GoMemoryLimit: bytes.Bytes{Value: configuredMem},
		ControllerNextGC: &nextgc.ControllerConfig{
			RSSLimit:             bytes.Bytes{Value: 1 << 30},
			DangerZoneGOGC:       50,
			DangerZoneThrottling: 90,
			Period:               duration.Duration{Duration: time.Hour},
			ComponentProportional: &nextgc.ComponentProportionalConfig{
				Coefficient: 1,
			},
		},
	}

	service, err := newServiceImpl(
		logger,
		cfg,
		&serviceStatsSubscriptionStub{},
		&backpressureOperatorStub{},
	)
	require.NoError(t, err)

	require.Equal(t, int64(configuredMem), debug.SetMemoryLimit(-1))

	service.Quit()

	require.Equal(t, initialLimit, debug.SetMemoryLimit(-1))
}
