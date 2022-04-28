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
		// нельзя потратить более 1 ГБ памяти (округлено до 1000 МБ для удобства)
		RSSLimit: bytes.Bytes{Value: 1000 * bytefmt.MEGABYTE},
		// при 50% исчерпании памяти начинается активное управление процессом
		DangerZoneGOGC: 50,
		// при 90% исчерпании памяти начинается активное управление процессом
		DangerZoneThrottling: 90,
		Period:               duration.Duration{Duration: controllerPeriod},
		ComponentProportional: &ComponentProportionalConfig{
			Coefficient: 1,
			WindowSize:  0, // выключаем сглаживание
		},
	}

	// Первый вариант статистики описывает ситуацию, когда память близка к исчерпанию
	memoryBudgetExhausted := &stats.Service{
		NextGC: 950 * bytefmt.MEGABYTE, // память потрачена на 95%
	}

	// Во втором варианте бюджет памяти возвращается в норму
	memoryBudgetNormal := &stats.Service{
		NextGC: 300 * bytefmt.MEGABYTE, // память потрачена на 50%
	}

	subscriptionMock := &stats.ServiceSubscriptionMock{
		Chan: make(chan *stats.Service),
	}

	// канал закрывается, когда backpressure.Operator получит все необходимые команды
	terminateChan := make(chan struct{})

	var serviceStatsContainer atomic.Value

	// имитация Servus - поставщика информации о статистике в канал
	go func() {
		ticker := time.NewTicker(servusPeriod)

		for {
			select {
			case <-ticker.C:
				serviceStats, ok := serviceStatsContainer.Load().(*stats.Service)
				if ok {
					subscriptionMock.Chan <- serviceStats
				}
			case <-terminateChan:
				return
			}
		}
	}()

	backpressureOperatorMock := &backpressure.OperatorMock{}

	// изначельно поставляется статистика, в которой память кажется исчерпанной
	serviceStatsContainer.Store(memoryBudgetExhausted)

	backpressureOperatorMock.On(
		"SetControlParameters",
		&stats.ControlParameters{
			GOGC:                 80,
			ThrottlingPercentage: 20,
		},
	).Return(nil).Once().Run(
		func(args mock.Arguments) {
			// после того, как управляющий сигнал на троттлинг был передан в backpressure,
			// подменяем статистику на такую, чтобы регулятор подумал, что ситуация с бюджетом памяти нормализовалась
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
