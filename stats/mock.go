package stats

import (
	"github.com/stretchr/testify/mock"
)

var _ Subscription = (*ServiceSubscriptionMock)(nil)

type ServiceSubscriptionMock struct {
	Chan chan *ServiceStats
	mock.Mock
}

func (m *ServiceSubscriptionMock) Updates() <-chan *ServiceStats {
	return m.Chan
}

func (m *ServiceSubscriptionMock) Quit() { m.Called() }
