package memlimiter

import (
	"gitlab.stageoffice.ru/UCS-COMMON/errors"

	"github.com/newcloudtechnologies/memlimiter/controller/nextgc"
)

// Config - высокоуровневая конфигурация ограничителя потребления памяти.
type Config struct {
	// TODO: при появлении новых имплементаций добавлять их сюда, в Prepare реализовать вариативность выбора
	// ControllerNextGC - настройки регулятора, управляющего бюджетом памяти
	ControllerNextGC *nextgc.ControllerConfig `json:"controller_nextgc"` //nolint:tagliatelle
}

// Prepare - валидатор конфига.
func (c *Config) Prepare() error {
	if c.ControllerNextGC == nil {
		return errors.New("empty ControllerNextGC")
	}

	return nil
}
