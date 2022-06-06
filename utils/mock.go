package utils

import (
	"github.com/stretchr/testify/mock"
)

var _ ConsumptionReporter = (*ConsumptionReporterMock)(nil)

type ConsumptionReporterMock struct {
	mock.Mock
}

func (m *ConsumptionReporterMock) PredefinedConsumers(serviceStats interface{}) (*ConsumptionReport, error) {
	// TODO implement me
	panic("implement me")
}

type ApplicationTerminatorMock struct {
	mock.Mock
}

func (m *ApplicationTerminatorMock) Terminate(fatalErr error) {
	// TODO implement me
	panic("implement me")
}
