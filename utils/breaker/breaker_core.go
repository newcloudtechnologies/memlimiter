package breaker

import (
	"runtime"
	"sync"
	"sync/atomic"

	"gitlab.stageoffice.ru/UCS-COMMON/errors"
)

// breakerCore предоставляет интерфейс для корректной остановки какой-либо подсистемы
type breakerCore struct {
	mutex *sync.RWMutex
	count int64
	mode  int32
}

// Inc инкрементирует счётчик активных задач
func (b *breakerCore) Inc() error {
	if !b.IsOperational() {
		return errors.New("shutdown in progress")
	}

	atomic.AddInt64(&b.count, 1)

	return nil
}

// IncN инкрементирует счётчик активных задач сразу на указанное число
func (b *breakerCore) IncN(n int) error {
	if !b.IsOperational() {
		return errors.New("shutdown in progress")
	}

	atomic.AddInt64(&b.count, int64(n))

	return nil
}

// IncWithLock инкрементирует счётчик активных задач и захватывает мьютекс
func (b *breakerCore) IncWithLock(exclusive bool) error {
	if exclusive {
		b.mutex.Lock()
	} else {
		b.mutex.RLock()
	}

	if err := b.Inc(); err != nil {
		// отпускаем блокировку, если не удалось инкрементировать счётчик
		if exclusive {
			b.mutex.Unlock()
		} else {
			b.mutex.RUnlock()
		}

		return err
	}

	return nil
}

// Dec декрементирует счётчик активных задач
func (b *breakerCore) Dec() {
	atomic.AddInt64(&b.count, -1)
}

// DecWithUnlock декрементирует счётчик активных задач и отпускает мьютекс
func (b *breakerCore) DecWithUnlock(exclusive bool) {
	b.Dec()

	if exclusive {
		b.mutex.Unlock()
	} else {
		b.mutex.RUnlock()
	}
}

// IsOperational проверяет, находится ли breakerCore в рабочем состоянии или уже всё
func (b *breakerCore) IsOperational() bool { return atomic.LoadInt32(&b.mode) == operational }

// Count возвращает текущее значение счетчика активных задач
func (b *breakerCore) Count() (int64, error) {
	if !b.IsOperational() {
		return 0, errors.New("shutdown in progress")
	}

	return atomic.LoadInt64(&b.count), nil
}

// Shutdown переводит выключатель в состояние завершения работа,
// в котором все вызовы метода Inc завершаются с ошибкой
func (b *breakerCore) Shutdown() {
	_ = atomic.CompareAndSwapInt32(&b.mode, operational, shutdown)
}

// Wait блокируется до момента обнуления счётчика активных задач
func (b *breakerCore) Wait() {
	if atomic.LoadInt32(&b.mode) != shutdown {
		panic("cannot wait on operational Breaker, turn it off first")
	}

	for {
		if atomic.LoadInt64(&b.count) == 0 {
			break
		}

		runtime.Gosched()
	}
}

func (b *breakerCore) ShutdownAndWait() {
	b.Shutdown()
	b.Wait()
}

func (b *breakerCore) Err() error {
	if b.IsOperational() {
		return nil
	}

	return errors.New("breaker is not operational")
}

// newBreakerCore создаёт новый "выключатель", который помогает подсчитывать
// активные действия в какой-то подсистеме и корректно завершать её работу
func newBreakerCore() *breakerCore {
	return &breakerCore{
		count: 0,
		mode:  operational,
	}
}
