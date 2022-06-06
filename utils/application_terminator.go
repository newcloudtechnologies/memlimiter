package utils

import (
	"os"

	"github.com/go-logr/logr"
)

// ApplicationTerminator - интерфейс для завершения работы приложения по команде MemLimiter.
// Применяется в тех ситуациях, когда лучше перезагрузить приложение, чем продолжать работу.
// Должен реализовываться пользователями библиотеки, так у каждого приложения свой протокол корректного завершения работы.
type ApplicationTerminator interface {
	// Terminate - специальный метод, регистрирующий фатальную ошибку управления бюджетом памяти.
	// Приложение обязано завершить работу, если этот метод был вызван хотя бы один раз.
	Terminate(fatalErr error)
}

type ungracefulApplicationTerminator struct {
	logger logr.Logger
}

func (at *ungracefulApplicationTerminator) Terminate(fatalErr error) {
	at.logger.Error(fatalErr, "terminate application due to fatal error")
	os.Exit(1)
}

// NewUngracefulApplicationTerminator создаёт тривиальную имплементацию выключателя,
// при поступлении ошибок процесс просто гасится с os.Exit(1).
// Использовать только для самых простых stateless сервисов.
func NewUngracefulApplicationTerminator(logger logr.Logger) ApplicationTerminator {
	return &ungracefulApplicationTerminator{
		logger: logger,
	}
}

type fatalErrChanApplicationTerminator struct {
	fatalErrChan chan<- error
}

func (at *fatalErrChanApplicationTerminator) Terminate(fatalErr error) { at.fatalErrChan <- fatalErr }

// NewFatalErrChanApplicationTerminator создаёт имплементацию выключателя,
// в котором ошибка записывается в специальный канал, и читатель из канала её может обработать особым образом.
func NewFatalErrChanApplicationTerminator(fatalErrChan chan<- error) ApplicationTerminator {
	return &fatalErrChanApplicationTerminator{
		fatalErrChan: fatalErrChan,
	}
}
