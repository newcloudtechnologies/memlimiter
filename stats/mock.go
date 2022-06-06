package stats

import (
	"github.com/stretchr/testify/mock"
)

var _ Subscription = (*ServiceSubscriptionMock)(nil)

type ServiceSubscriptionMock struct {
	Chan chan *Service
	mock.Mock
}

func (m *ServiceSubscriptionMock) Updates() <-chan *Service {
	return m.Chan
}

func (m *ServiceSubscriptionMock) Quit() { m.Called() }
