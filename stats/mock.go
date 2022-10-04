/*
 * Copyright (c) New Cloud Technologies, Ltd. 2013-2022.
 * Author: Vitaly Isaev <vitaly.isaev@myoffice.team>
 * License: https://github.com/newcloudtechnologies/memlimiter/blob/master/LICENSE
 */

package stats

import (
	"github.com/stretchr/testify/mock"
)

var _ ServiceStats = (*ServiceStatsMock)(nil)

// ServiceStatsMock mocks ServiceStatsSubscription.
type ServiceStatsMock struct {
	mock.Mock
}

func (m *ServiceStatsMock) RSS() uint64 {
	return m.Called().Get(0).(uint64)
}

func (m *ServiceStatsMock) NextGC() uint64 {
	return m.Called().Get(0).(uint64)
}

func (m *ServiceStatsMock) ConsumptionReport() *ConsumptionReport {
	args := m.Called()

	return args.Get(0).(*ConsumptionReport)
}

var _ ServiceStatsSubscription = (*ServiceStatsSubscriptionMock)(nil)

// ServiceStatsSubscriptionMock mocks ServiceStatsSubscription.
type ServiceStatsSubscriptionMock struct {
	ServiceStatsSubscription
	Chan chan ServiceStats
	mock.Mock
}

func (m *ServiceStatsSubscriptionMock) Updates() <-chan ServiceStats {
	return m.Chan
}
