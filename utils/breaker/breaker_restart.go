package breaker

import "sync/atomic"

// BreakerRestart предоставляет интерфейс для корректной остановки какой-либо подсистемы
type BreakerRestart struct {
	*breakerCore
}

// Reset переводит обратно в рабочий режим
func (b *BreakerRestart) Reset() {
	if !atomic.CompareAndSwapInt32(&b.mode, shutdown, operational) {
		panic("cannot reset Breaker, turn it off first")
	}
}

// NewBreakerRestart расширяет функционал брейкера перезапуском внутреннего счетчика
func NewBreakerRestart() *BreakerRestart {
	return &BreakerRestart{
		breakerCore: newBreakerCore(),
	}
}
