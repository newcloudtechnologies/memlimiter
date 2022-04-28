package backpressure

import (
	"github.com/newcloudtechnologies/memlimiter/stats"
	"github.com/stretchr/testify/mock"
)

var _ Operator = (*OperatorMock)(nil)

type OperatorMock struct {
	mock.Mock
}

func (m *OperatorMock) GetStats() *stats.Backpressure {
	// TODO implement me
	panic("implement me")
}

func (m *OperatorMock) SetControlParameters(value *stats.ControlParameters) error {
	args := m.Called(value)

	return args.Error(0)
}

func (m *OperatorMock) AllowRequest() bool {
	// TODO implement me
	panic("implement me")
}
