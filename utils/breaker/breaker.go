package breaker

import (
	"sync"
	"sync/atomic"
	"time"
)

const (
	operational int32 = iota + 1
	shutdown
)

// Breaker предоставляет интерфейс для корректной остановки какой-либо подсистемы
type Breaker struct {
	*breakerCore
	exitChan chan struct{}
}

// Shutdown переводит выключатель в состояние завершения работы,
// в котором все вызовы метода Inc завершаются с ошибкой
func (b *Breaker) Shutdown() {
	if atomic.CompareAndSwapInt32(&b.mode, operational, shutdown) {
		close(b.exitChan)
	}
}

func (b *Breaker) ShutdownAndWait() {
	b.Shutdown()
	b.Wait()
}

func (b *Breaker) Deadline() (deadline time.Time, ok bool) {
	return time.Time{}, false
}

func (b *Breaker) Value(key interface{}) interface{} { return nil }

// Done возвращает канал, который можно опрашивать на предмет завершённости работы
func (b *Breaker) Done() <-chan struct{} { return b.exitChan }

// NewBreaker создаёт новый "выключатель", который помогает подсчитывать
// активные действия в какой-то подсистеме и корректно завершать её работу
func NewBreaker() *Breaker {
	return &Breaker{
		breakerCore: newBreakerCore(),
		exitChan:    make(chan struct{}),
	}
}

// NewBreakerWithInitValue - альтернативный конструктор выключателя, который удобно применять при
// создании акторов, когда уже точно известно, сколько горутин запускается при старте
func NewBreakerWithInitValue(value int64) *Breaker {
	b := NewBreaker()
	b.count = value

	return b
}

// NewBreakerWithMutex добавляет в выключатель мьютекс, чтобы можно было пользоваться методами IncWithLock / DecWithUnlock
func NewBreakerWithMutex() *Breaker {
	b := NewBreaker()
	b.mutex = &sync.RWMutex{}

	return b
}
