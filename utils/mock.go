package utils

import (
	"github.com/stretchr/testify/mock"

	"gitlab.stageoffice.ru/UCS-COMMON/schemagen-go/v41/servus/stats/v1"
)

var _ ConsumptionReporter = (*ConsumptionReporterMock)(nil)

type ConsumptionReporterMock struct {
	mock.Mock
}

func (m *ConsumptionReporterMock) PredefinedConsumers(serviceStats *stats.ServiceStats) (*ConsumptionReport, error) {
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
