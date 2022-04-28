package nextgc

import (
	"sync/atomic"
	"testing"
	"time"

	"code.cloudfoundry.org/bytefmt"
	"github.com/go-logr/logr/testr"
	servus_stats "gitlab.stageoffice.ru/UCS-COMMON/schemagen-go/v41/servus/stats/v1"

	"github.com/newcloudtechnologies/memlimiter/utils/config/bytes"
	"github.com/newcloudtechnologies/memlimiter/utils/config/duration"
	"github.com/stretchr/testify/mock"
	"gitlab.stageoffice.ru/UCS-PLATFORM/servus/stats/aggregate"

	"github.com/newcloudtechnologies/memlimiter/backpressure"
	"github.com/newcloudtechnologies/memlimiter/utils"
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
	memoryBudgetExhausted := &servus_stats.ServiceStats{
		Process: &servus_stats.ProcessStats{
			LanguageSpecific: &servus_stats.ProcessStats_Go{
				Go: &servus_stats.GoRuntimeStats{
					MemStats: &servus_stats.GoMemStats{
						NextGc: 950 * bytefmt.MEGABYTE, // память потрачена на 95%
					},
				},
			},
		},
	}

	// Во втором варианте бюджет памяти возвращается в норму
	memoryBudgetNormal := &servus_stats.ServiceStats{
		Process: &servus_stats.ProcessStats{
			LanguageSpecific: &servus_stats.ProcessStats_Go{
				Go: &servus_stats.GoRuntimeStats{
					MemStats: &servus_stats.GoMemStats{
						NextGc: 300 * bytefmt.MEGABYTE, // память потрачена на 50%
					},
				},
			},
		},
	}

	subscriptionMock := &aggregate.SubscriptionMock{}
	updateChan := make(chan *servus_stats.ServiceStats)
	subscriptionMock.On("Updates").Return((<-chan *servus_stats.ServiceStats)(updateChan)) //nolint:gocritic

	// канал закрывается, когда backpressure.Operator получит все необходимые команды
	terminateChan := make(chan struct{})

	var serviceStatsContainer atomic.Value

	// имитация Servus - поставщика информации о статистике в канал
	go func() {
		ticker := time.NewTicker(servusPeriod)

		for {
			select {
			case <-ticker.C:
				serviceStats, ok := serviceStatsContainer.Load().(*servus_stats.ServiceStats)
				if ok {
					updateChan <- serviceStats
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
		&backpressure.ControlParameters{
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
		&backpressure.ControlParameters{
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
