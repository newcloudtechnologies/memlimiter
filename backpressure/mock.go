package backpressure

import (
	"github.com/stretchr/testify/mock"

	"gitlab.stageoffice.ru/UCS-COMMON/schemagen-go/v41/servus/stats/v1"
)

var _ Operator = (*OperatorMock)(nil)

type OperatorMock struct {
	mock.Mock
}

func (m *OperatorMock) GetStats() *stats.GoMemLimiterStats_BackpressureStats {
	// TODO implement me
	panic("implement me")
}

func (m *OperatorMock) SetControlParameters(value *ControlParameters) error {
	args := m.Called(value)

	return args.Error(0)
}

func (m *OperatorMock) AllowRequest() bool {
	// TODO implement me
	panic("implement me")
}
