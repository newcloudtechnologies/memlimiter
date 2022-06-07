package stats

import (
	"github.com/stretchr/testify/mock"
)

var _ ServiceStats = (*ServiceStatsMock)(nil)

// ServiceStatsMock mocks ServiceStatsSubscription.
type ServiceStatsMock struct {
	mock.Mock
}

func (m *ServiceStatsMock) NextGC() uint64 {
	return m.Called().Get(0).(uint64)
}

func (m *ServiceStatsMock) PredefinedConsumers() (*ConsumptionReport, error) {
	args := m.Called()

	return args.Get(0).(*ConsumptionReport), args.Error(1)
}

var _ ServiceStatsSubscription = (*SubscriptionMock)(nil)

// SubscriptionMock mocks ServiceStatsSubscription.
type SubscriptionMock struct {
	Chan chan ServiceStats
	mock.Mock
}

func (m *SubscriptionMock) Updates() <-chan ServiceStats {
	return m.Chan
}

func (m *SubscriptionMock) Quit() { m.Called() }
