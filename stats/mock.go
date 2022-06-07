package stats

import (
	"github.com/stretchr/testify/mock"
)

var _ Subscription = (*SubscriptionMock)(nil)

// SubscriptionMock mocks Subscription
type SubscriptionMock struct {
	Chan chan *ServiceStats
	mock.Mock
}

func (m *SubscriptionMock) Updates() <-chan *ServiceStats {
	return m.Chan
}

func (m *SubscriptionMock) Quit() { m.Called() }
