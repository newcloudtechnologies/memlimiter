package nextgc

import (
	"math"
	"time"

	"github.com/go-logr/logr"
	"github.com/newcloudtechnologies/memlimiter/stats"
	"github.com/newcloudtechnologies/memlimiter/utils/breaker"
	"github.com/pkg/errors"

	"github.com/newcloudtechnologies/memlimiter/backpressure"
	"github.com/newcloudtechnologies/memlimiter/controller"
	memlimiter_utils "github.com/newcloudtechnologies/memlimiter/utils"
)

// controllerImpl - ПД-регулятор, описанный в
// https://confluence.ncloudtech.ru/pages/viewpage.action?pageId=102850263&src=contextnavpagetreemode
//nolint:govet // структура создаётся в ед. экземпляре, сильно ни на что не влияет
type controllerImpl struct {
	input                 stats.Subscription                     // вход: поток оперативной статистики сервиса
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
	controlParameters *stats.ControlParameters            // значение управляющих параметров

	getStatsChan chan *getStatsRequest

	cfg     *ControllerConfig
	logger  logr.Logger
	breaker *breaker.Breaker
}

type getStatsRequest struct {
	result chan *stats.Controller
}

func (r *getStatsRequest) respondWith(resp *stats.Controller) {
	r.result <- resp
}

func (c *controllerImpl) GetStats() (*stats.Controller, error) {
	req := &getStatsRequest{result: make(chan *stats.Controller, 1)}

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
				c.logger.Error(err, "update state")
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
				c.logger.Error(err, "apply control value")
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

func (c *controllerImpl) updateState(serviceStats *stats.Service) error {
	// извлекаем оперативную информацию о спец. потребителях памяти, если она предоставляется клиентом
	if c.consumptionReporter != nil {
		var err error

		c.consumptionReport, err = c.consumptionReporter.PredefinedConsumers(serviceStats.Custom)
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

func (c *controllerImpl) updateUtilization(serviceStats *stats.Service) {
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

	c.utilization = float64(serviceStats.NextGC) / float64(c.goAllocLimit)
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
	c.controlParameters = &stats.ControlParameters{}
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
		c.logger.Info("control parameters are not ready yet")

		return nil
	}

	if err := c.backpressureOperator.SetControlParameters(c.controlParameters); err != nil {
		return errors.Wrapf(err, "set control parameters: %v", c.controlParameters)
	}

	return nil
}

func (c *controllerImpl) aggregateStats() *stats.Controller {
	res := &stats.Controller{
		MemoryBudget: &stats.MemoryBudget{
			RSSLimit:     c.cfg.RSSLimit.Value,
			GoAllocLimit: c.goAllocLimit,
			Utilization:  c.utilization,
		},
		NextGC: &stats.ControllerNextGC{
			P:      c.pValue,
			Output: c.sumValue,
		},
	}

	if c.consumptionReport != nil {
		res.MemoryBudget.SpecialConsumers = &stats.SpecialConsumers{}
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
	logger logr.Logger,
	cfg *ControllerConfig,
	input stats.Subscription,
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
