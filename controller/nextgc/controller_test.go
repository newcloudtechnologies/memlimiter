package nextgc

import (
	"sync/atomic"
	"testing"
	"time"

	"code.cloudfoundry.org/bytefmt"
	"github.com/go-logr/logr/testr"
	"github.com/newcloudtechnologies/memlimiter/stats"

	"github.com/newcloudtechnologies/memlimiter/backpressure"
	"github.com/newcloudtechnologies/memlimiter/utils"
	"github.com/newcloudtechnologies/memlimiter/utils/config/bytes"
	"github.com/newcloudtechnologies/memlimiter/utils/config/duration"
	"github.com/stretchr/testify/mock"
)

func TestController(t *testing.T) {
	logger := testr.New(t)

	const servusPeriod = 100 * time.Millisecond

	controllerPeriod := 2 * servusPeriod

	cfg := &ControllerConfig{
		// We cannot exceed 1000M RSS threshold
		RSSLimit: bytes.Bytes{Value: 1000 * bytefmt.MEGABYTE},
		// When memory budget utilization reaches 50%, the controller will start GOGC altering.
		DangerZoneGOGC: 50,
		// When memory budget utilization reaches 90%, the controller will start request throttling.
		DangerZoneThrottling: 90,
		Period:               duration.Duration{Duration: controllerPeriod},
		ComponentProportional: &ComponentProportionalConfig{
			Coefficient: 1,
			WindowSize:  0, // just for simplicity disable the smoothing
		},
	}

	// First ServiceStats instance describes the situation, when the memory budget utilization
	// is very close to the limits.
	memoryBudgetExhausted := &stats.ServiceStats{
		NextGC: 950 * bytefmt.MEGABYTE,
	}

	// In the second case the memory budget utilization returns to the ordinary values.
	memoryBudgetNormal := &stats.ServiceStats{
		NextGC: 300 * bytefmt.MEGABYTE, // память потрачена на 50%
	}

	subscriptionMock := &stats.SubscriptionMock{
		Chan: make(chan *stats.ServiceStats),
	}

	// this channel is closed when backpressure.Operator receives all required actions
	terminateChan := make(chan struct{})

	var serviceStatsContainer atomic.Value

	// The stream of stats.ServiceStats instances
	go func() {
		ticker := time.NewTicker(servusPeriod)

		for {
			select {
			case <-ticker.C:
				serviceStats, ok := serviceStatsContainer.Load().(*stats.ServiceStats)
				if ok {
					subscriptionMock.Chan <- serviceStats
				}
			case <-terminateChan:
				return
			}
		}
	}()

	backpressureOperatorMock := &backpressure.OperatorMock{}

	// Here we model the situation of memory exhaustion.
	serviceStatsContainer.Store(memoryBudgetExhausted)

	backpressureOperatorMock.On(
		"SetControlParameters",
		&stats.ControlParameters{
			GOGC:                 80,
			ThrottlingPercentage: 20,
		},
	).Return(nil).Once().Run(
		func(args mock.Arguments) {
			// As soon as the control signal is delivered to the backpressure.Operator,
			// replace the ServiceStats instance to make controller think that memory
			// consumption returned to normal.
			serviceStatsContainer.Store(memoryBudgetNormal)
		},
	).On(
		"SetControlParameters",
		&stats.ControlParameters{
			GOGC:                 100,
			ThrottlingPercentage: 0,
		},
	).Return(nil).Once().Run(
		func(args mock.Arguments) {
			close(terminateChan)
		},
	)

	consumptionReporterMock := &utils.ConsumptionReporterMock{}

	c := NewControllerFromConfig(logger, cfg, subscriptionMock, nil, backpressureOperatorMock, &utils.ApplicationTerminatorMock{})

	<-terminateChan

	c.Quit()

	mock.AssertExpectationsForObjects(t, subscriptionMock, backpressureOperatorMock, consumptionReporterMock)
}
