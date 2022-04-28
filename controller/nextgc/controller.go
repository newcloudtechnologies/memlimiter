package nextgc

import (
	"math"
	"time"

	servus_stats "gitlab.stageoffice.ru/UCS-COMMON/schemagen-go/v41/servus/stats/v1"

	"github.com/pkg/errors"
	"gitlab.stageoffice.ru/UCS-COMMON/gaben"
	"github.com/newcloudtechnologies/memlimiter/utils/breaker"
	"gitlab.stageoffice.ru/UCS-PLATFORM/servus/stats/aggregate"

	"github.com/newcloudtechnologies/memlimiter/backpressure"
	"github.com/newcloudtechnologies/memlimiter/controller"
	memlimiter_utils "github.com/newcloudtechnologies/memlimiter/utils"
)

// controllerImpl - ПД-регулятор, описанный в
// https://confluence.ncloudtech.ru/pages/viewpage.action?pageId=102850263&src=contextnavpagetreemode
//nolint:govet // структура создаётся в ед. экземпляре, сильно ни на что не влияет
type controllerImpl struct {
	input                 aggregate.Subscription                 // вход: поток оперативной статистики сервиса
	consumptionReporter   memlimiter_utils.ConsumptionReporter   // вход: информация о специализированных потребителях памяти
	backpressureOperator  backpressure.Operator                  // выход: обрабатывает управляющие команды, выдаваемые регулятором
	applicationTerminator memlimiter_utils.ApplicationTerminator // выход: аварийная остановка приложения

	// компоненты регулятора
	//
	// NOTE: На определённом этапе планировалось сделать полноценный ПИД-регулятор, но в итоге вышел
	// не совсем обычный ПД-регулятор. В пропорциональной составляющей заложена
	// не прямая, а обратная пропорциональность, а интегральная и дифференциальная составляющие отсутствуют.
	// Интегральной, видимо, не будет никогда, а дифференциальную нужно будет доработать,
	// если в системе будут наблюдаться проблемы с автоколебаниями.
	componentP *componentP

	// закешированные значения, описывающее актуальное состояние ПИД-регулятора
	//
	// NOTE: см. замечание выше
	pValue   float64 // значение, выданное пропорциональным компонентом
	sumValue float64 // итоговое значение управляющего сигнала

	goAllocLimit      uint64                              // акутальный бюджет памяти для Go [байты]
	utilization       float64                             // актуальная утилизация бюджета памяти
	consumptionReport *memlimiter_utils.ConsumptionReport // отчёт о потреблении памяти специальными потребителями
	controlParameters *backpressure.ControlParameters     // значение управляющих параметров

	getStatsChan chan *getStatsRequest

	cfg     *ControllerConfig
	logger  gaben.Logger
	breaker *breaker.Breaker
}

type getStatsRequest struct {
	result chan *servus_stats.GoMemLimiterStats_ControllerStats
}

func (r *getStatsRequest) respondWith(stats *servus_stats.GoMemLimiterStats_ControllerStats) {
	r.result <- stats
}

func (c *controllerImpl) GetStats() (*servus_stats.GoMemLimiterStats_ControllerStats, error) {
	req := &getStatsRequest{result: make(chan *servus_stats.GoMemLimiterStats_ControllerStats, 1)}

	select {
	case c.getStatsChan <- req:
	case <-c.breaker.Done():
		return nil, c.breaker.Err()
	}

	select {
	case resp := <-req.result:
		return resp, nil
	case <-c.breaker.Done():
		return nil, c.breaker.Err()
	}
}

func (c *controllerImpl) loop() {
	defer c.breaker.Dec()

	ticker := time.NewTicker(c.cfg.Period.Duration)
	defer ticker.Stop()

	for {
		select {
		case serviceStats := <-c.input.Updates():
			// обновление состояния осуществляется при каждом поступлении статистики от Сервуса
			if err := c.updateState(serviceStats); err != nil {
				c.logger.Error("update state", gaben.Error(err))
				// невозможность вычислить контрольные параметры - фатальная ошибка;
				// завершаем приложение через побочный эффект
				c.applicationTerminator.Terminate(err)

				return
			}
		case <-ticker.C:
			// а вот отправка управляющего сигнала выполняется с периодичностью, независящей от Сервуса;
			// клиент обязан сам убедиться, что период отправки управляющего сигнала больше либо равен периоду отправки статистики Сервусом,
			// иначе один и тот же сигнал будет высылаться несколько раз
			if err := c.applyControlValue(); err != nil {
				c.logger.Error("apply control value", gaben.Error(err))
				// невозможность применить контрольные параметры - фатальная ошибка;
				// завершаем приложение через побочный эффект
				c.applicationTerminator.Terminate(err)

				return
			}
		case req := <-c.getStatsChan:
			req.respondWith(c.aggregateStats())
		case <-c.breaker.Done():
			return
		}
	}
}

func (c *controllerImpl) updateState(serviceStats *servus_stats.ServiceStats) error {
	// извлекаем оперативную информацию о спец. потребителях памяти, если она предоставляется клиентом
	if c.consumptionReporter != nil {
		var err error

		c.consumptionReport, err = c.consumptionReporter.PredefinedConsumers(serviceStats)
		if err != nil {
			return errors.Wrap(err, "predefined consumers")
		}
	}

	c.updateUtilization(serviceStats)

	if err := c.updateControlValues(); err != nil {
		return errors.Wrap(err, "update control values")
	}

	c.updateControlParameters()

	return nil
}

func (c *controllerImpl) updateUtilization(serviceStats *servus_stats.ServiceStats) {
	// Чтобы понять, сколько памяти можно аллоцировать на нужды Go,
	// требуется вычесть из общего лимита на RSS память, в явном виде потраченную в CGO.
	// Иными словами, если аллокации в CGO будут расти, то аллокации в Go должны ужиматься.
	var cgoAllocs uint64

	if c.consumptionReport != nil {
		for _, value := range c.consumptionReport.Cgo {
			cgoAllocs += value
		}
	}

	c.goAllocLimit = c.cfg.RSSLimit.Value - cgoAllocs

	nextGC := serviceStats.Process.GetGo().MemStats.NextGc
	c.utilization = float64(nextGC) / float64(c.goAllocLimit)
}

func (c *controllerImpl) updateControlValues() error {
	var err error

	c.pValue, err = c.componentP.value(c.utilization)
	if err != nil {
		return errors.Wrap(err, "component p value")
	}

	// TODO: при появлении новых компонент суммировать их значения здесь:
	c.sumValue = c.pValue

	// Сатурация выхода регулятора, чтобы эффект вышел не слишком сильным
	// Детали:
	// https://en.wikipedia.org/wiki/Saturation_arithmetic
	// https://habr.com/ru/post/345972/
	const (
		lowerBound = 0
		upperBound = 99 // иначе GOGC превратится в 0
	)

	c.sumValue = memlimiter_utils.ClampFloat64(c.sumValue, lowerBound, upperBound)

	return nil
}

func (c *controllerImpl) updateControlParameters() {
	c.controlParameters = &backpressure.ControlParameters{}
	c.updateControlParameterGOGC()
	c.updateControlParameterThrottling()
}

const percents = 100

func (c *controllerImpl) updateControlParameterGOGC() {
	// 	В зелёной зоне воздействие сбрасывается до дефолтов.
	if uint32(c.utilization*percents) < c.cfg.DangerZoneGOGC {
		c.controlParameters.GOGC = backpressure.DefaultGOGC

		return
	}

	// В красной зоне для GC устанавливаются более консервативные настройки
	roundedValue := uint32(math.Round(c.sumValue))
	c.controlParameters.GOGC = int(backpressure.DefaultGOGC - roundedValue)
}

func (c *controllerImpl) updateControlParameterThrottling() {
	// 	В зелёной зоне воздействие сбрасывается до дефолтов.
	if uint32(c.utilization*percents) < c.cfg.DangerZoneThrottling {
		c.controlParameters.ThrottlingPercentage = backpressure.NoThrottling

		return
	}

	// В красной зоне для GC устанавливаются более консервативные настройки
	roundedValue := uint32(math.Round(c.sumValue))
	c.controlParameters.ThrottlingPercentage = roundedValue
}

func (c *controllerImpl) applyControlValue() error {
	if c.controlParameters == nil {
		c.logger.Warning("control parameters are not ready yet")

		return nil
	}

	if err := c.backpressureOperator.SetControlParameters(c.controlParameters); err != nil {
		return errors.Wrapf(err, "set control parameters: %v", c.controlParameters)
	}

	return nil
}

func (c *controllerImpl) aggregateStats() *servus_stats.GoMemLimiterStats_ControllerStats {
	res := &servus_stats.GoMemLimiterStats_ControllerStats{
		MemoryBudget: &servus_stats.GoMemLimiterStats_ControllerStats_MemoryBudget{
			RssLimit:     c.cfg.RSSLimit.Value,
			GoAllocLimit: c.goAllocLimit,
			Utilization:  c.utilization,
		},
		Controller: &servus_stats.GoMemLimiterStats_ControllerStats_NextGc{
			NextGc: &servus_stats.GoMemLimiterStats_ControllerStats_ControllerNextGC{
				P:      c.pValue,
				Output: c.sumValue,
			},
		},
	}

	if c.consumptionReport != nil {
		res.MemoryBudget.SpecialConsumers = &servus_stats.GoMemLimiterStats_ControllerStats_MemoryBudget_SpecialConsumers{}
		res.MemoryBudget.SpecialConsumers.Go = c.consumptionReport.Go
		res.MemoryBudget.SpecialConsumers.Cgo = c.consumptionReport.Cgo
	}

	return res
}

// Quit корректно завершает работу.
func (c *controllerImpl) Quit() {
	c.breaker.Shutdown()
	c.breaker.Wait()
}

// NewControllerFromConfig - контроллер регулятора.
func NewControllerFromConfig(
	logger gaben.Logger,
	cfg *ControllerConfig,
	input aggregate.Subscription,
	consumptionReporter memlimiter_utils.ConsumptionReporter,
	backpressureOperator backpressure.Operator,
	applicationTerminator memlimiter_utils.ApplicationTerminator,
) controller.Controller {
	c := &controllerImpl{
		input:                 input,
		consumptionReporter:   consumptionReporter,
		backpressureOperator:  backpressureOperator,
		componentP:            newComponentP(logger, cfg.ComponentProportional),
		pValue:                0,
		sumValue:              0,
		applicationTerminator: applicationTerminator,
		controlParameters:     nil,
		getStatsChan:          make(chan *getStatsRequest),
		cfg:                   cfg,
		logger:                logger,
		breaker:               breaker.NewBreakerWithInitValue(1),
	}

	go c.loop()

	return c
}
