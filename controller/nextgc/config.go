package nextgc

import (
	"gitlab.stageoffice.ru/UCS-COMMON/errors"
	"gitlab.stageoffice.ru/UCS-COMMON/utils/config/bytes"
	"gitlab.stageoffice.ru/UCS-COMMON/utils/config/duration"
)

// ControllerConfig - конфигурация PD-регулятора.
type ControllerConfig struct {
	// RSSLimit - верхний предел потребления RSS процессом
	RSSLimit bytes.Bytes `json:"rss_limit"`
	// DangerZoneGOGC - пороговое значение утилизации памяти, при котором регулятор начинает
	// устанавливать более консервативные настройки для GC
	// допустимые значения - (0; 100)
	DangerZoneGOGC uint32 `json:"danger_zone_gogc"`
	// DangerZoneThrottling - пороговое значение утилизации памяти, при котором регулятор начинает
	// подавлять пользовательские запросы
	// допустимые значения - (0; 100)
	DangerZoneThrottling uint32 `json:"danger_zone_throttling"`
	// Period - периодичность генерации управляющих сигналов
	Period duration.Duration `json:"period"`
	// ComponentProportional - конфигурация пропорциональной составляющей регулятора
	ComponentProportional *ComponentProportionalConfig `json:"component_proportional"`
}

// Prepare - валидатор конфига.
func (c *ControllerConfig) Prepare() error {
	if c.RSSLimit.Value == 0 {
		return errors.New("empty RSSLimit")
	}

	if c.DangerZoneGOGC == 0 || c.DangerZoneGOGC > 100 {
		return errors.Newf("invalid DangerZoneGOGC value (must belong to [0; 100])")
	}

	if c.DangerZoneThrottling == 0 || c.DangerZoneThrottling > 100 {
		return errors.Newf("invalid DangerZoneThrottling value (must belong to [0; 100])")
	}

	if c.Period.Duration == 0 {
		return errors.New("empty Period")
	}

	if c.ComponentProportional == nil {
		return errors.New("empty ComponentProportional")
	}

	return nil
}

// ComponentProportionalConfig - конфигурация пропорциональной составляющей регулятора.
type ComponentProportionalConfig struct {
	// Coefficient - коэффициент для вычисления взвешенной суммы в уравнении PID-регулятора
	Coefficient float64 `json:"coefficient"`
	// WindowSize - окно осреднения для экспоненциального скользящего среднего;
	// если равно нулю, то осреднение не выполняется.
	WindowSize uint `json:"window_size"`
}

// Prepare - валидатор конфига.
func (c *ComponentProportionalConfig) Prepare() error {
	if c.Coefficient == 0 {
		return errors.Newf("empty Coefficient makes no sense")
	}

	return nil
}
