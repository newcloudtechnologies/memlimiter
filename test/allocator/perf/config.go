package perf

import (
	"golang.org/x/time/rate"

	"github.com/newcloudtechnologies/memlimiter/utils/config/bytes"
	"github.com/newcloudtechnologies/memlimiter/utils/config/duration"
	"github.com/pkg/errors"
)

// Config - конфигурация нагрузочного клиента.
type Config struct {
	Endpoint       string            `json:"endpoint"`        // адрес сервера
	RPS            rate.Limit        `json:"rps"`             // количество запросов к сервису в секунду
	LoadDuration   duration.Duration `json:"load_duration"`   // продолжительность тестовой сессии
	AllocationSize bytes.Bytes       `json:"allocation_size"` // размер аллокации в каждом запросе
	PauseDuration  duration.Duration `json:"pause_duration"`  // пауза в хендлере сервиса (чтобы аллокация подольше продержались в памяти)
	RequestTimeout duration.Duration `json:"request_timeout"` // таймаут запроса к сервису
}

// Prepare - валидатор конфига.
func (c *Config) Prepare() error {
	if c.Endpoint == "" {
		return errors.New("empty endpoint")
	}

	if c.RPS == 0 {
		return errors.New("empty rps")
	}

	if c.LoadDuration.Duration == 0 {
		return errors.New("empty load duration")
	}

	if c.AllocationSize.Value == 0 {
		return errors.New("empty allocation size")
	}

	if c.PauseDuration.Duration == 0 {
		return errors.New("empty pause duration")
	}

	if c.RequestTimeout.Duration == 0 {
		return errors.New("empty request timeout")
	}

	return nil
}
