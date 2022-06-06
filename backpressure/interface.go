package backpressure

import (
	"github.com/newcloudtechnologies/memlimiter/stats"
)

const (
	// DefaultGOGC - значение GOGC по умолчанию.
	DefaultGOGC = 100
	// NoThrottling - разрешаем исполнение всех запросов.
	NoThrottling = 0
	// FullThrottling - полный запрет на исполнение запросов.
	FullThrottling = 100
)

// Operator приводит в действие управляющие сигналы.
type Operator interface {
	// SetControlParameters регистрирует актуальное значение управляющих параметров
	SetControlParameters(value *stats.ControlParameters) error
	// AllowRequest используется интерсепторами запросов для подавления части запросов во время пиковых нагрузок.
	AllowRequest() bool
	// GetStats возвращает статистику подсистемы backpressure
	GetStats() *stats.Backpressure
}
