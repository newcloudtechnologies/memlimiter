package utils

import (
	go_metrics "github.com/rcrowley/go-metrics"
)

var _ Counter = (*childCounter)(nil)

// Counter - абстракция счётчика
type Counter interface {
	go_metrics.Counter
}

// childCounter позволяет строить иерархические счётчики
type childCounter struct {
	Counter
	parent Counter
}

func (counter *childCounter) Dec(i int64) {
	counter.parent.Dec(i)
	counter.Counter.Dec(i)
}

func (counter *childCounter) Inc(i int64) {
	counter.parent.Inc(i)
	counter.Counter.Inc(i)
}

// NewCounter создаёт счетчик, ссылающийся на родительский счётчик.
// Если передать nil, то получится корневой счётчик в иерархии или просто отдельно стоящий счётчик.
func NewCounter(parent Counter) Counter {
	if parent == nil {
		return go_metrics.NewCounter()
	}

	return &childCounter{Counter: go_metrics.NewCounter(), parent: parent}
}
