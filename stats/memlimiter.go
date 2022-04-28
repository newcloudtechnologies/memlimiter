package stats

import (
	"fmt"
)

type Memlimiter struct {
	Controller   *Controller
	Backpressure *Backpressure
}

type Controller struct {
	MemoryBudget *MemoryBudget
	NextGC       *ControllerNextGC
}

type MemoryBudget struct {
	RSSLimit         uint64
	GoAllocLimit     uint64
	Utilization      float64
	SpecialConsumers *SpecialConsumers
}

type SpecialConsumers struct {
	Go  map[string]uint64
	Cgo map[string]uint64
}

type ControllerNextGC struct {
	P      float64
	Output float64
}

type Backpressure struct {
	Throttling        *Throttling
	ControlParameters *ControlParameters
}

type Throttling struct {
	Total     uint64
	Passed    uint64
	Throttled uint64
}

// ControlParameters - вектор управляющих сигналов для системы.
type ControlParameters struct {
	GOGC                 int    // значение GOGC (принимаются значения в формате debug.SetGCPercent)
	ThrottlingPercentage uint32 // процент запросов, которые должны быть отсечены на входе в сервис (в диапазоне [0; 100])
}

func (cp *ControlParameters) String() string {
	return fmt.Sprintf("gogc = %v, throttling_percentage = %v", cp.GOGC, cp.ThrottlingPercentage)
}

// ToKeysAndValues serializes struct for use in logr.Logger
func (cp *ControlParameters) ToKeysAndValues() []interface{} {
	return []interface{}{
		"gogc", cp.GOGC,
		"throttling_percentage", cp.ThrottlingPercentage,
	}
}

func (cp *ControlParameters) EqualsTo(other *ControlParameters) bool {
	return cp.GOGC == other.GOGC && cp.ThrottlingPercentage == other.ThrottlingPercentage
}
