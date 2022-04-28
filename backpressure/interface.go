package backpressure

import (
	"fmt"

	servus_stats "gitlab.stageoffice.ru/UCS-COMMON/schemagen-go/v41/servus/stats/v1"

	"gitlab.stageoffice.ru/UCS-COMMON/gaben"
)

// ControlParameters - вектор управляющих сигналов для системы.
type ControlParameters struct {
	GOGC                 int    // значение GOGC (принимаются значения в формате debug.SetGCPercent)
	ThrottlingPercentage uint32 // процент запросов, которые должны быть отсечены на входе в сервис (в диапазоне [0; 100])
}

func (cp *ControlParameters) String() string {
	return fmt.Sprintf("gogc = %v, throttling_percentage = %v", cp.GOGC, cp.ThrottlingPercentage)
}

// ToGaben конвертирует структуру для печати в логах.
func (cp *ControlParameters) ToGaben() []gaben.Field {
	return []gaben.Field{
		gaben.Int("gogc", cp.GOGC),
		gaben.Uint32("throttling_percentage", cp.ThrottlingPercentage),
	}
}

// ToProtobuf конвертирует структуру в соответствующий тип схемы.
func (cp *ControlParameters) ToProtobuf() *servus_stats.GoMemLimiterStats_BackpressureStats_ControlParameters {
	return &servus_stats.GoMemLimiterStats_BackpressureStats_ControlParameters{
		Gogc:                 uint32(cp.GOGC),
		ThrottlingPercentage: cp.ThrottlingPercentage,
	}
}

func (cp *ControlParameters) equalsTo(other *ControlParameters) bool {
	return cp.GOGC == other.GOGC && cp.ThrottlingPercentage == other.ThrottlingPercentage
}

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
	SetControlParameters(value *ControlParameters) error
	// AllowRequest используется интерсепторами запросов для подавления части запросов во время пиковых нагрузок.
	AllowRequest() bool
	// GetStats возвращает статистику подсистемы backpressure
	GetStats() *servus_stats.GoMemLimiterStats_BackpressureStats
}
