package memlimiter

import (
	"google.golang.org/grpc"

	"gitlab.stageoffice.ru/UCS-PLATFORM/servus"
)

// MemLimiter - верхнеуровневый интерфейс системы управления бюджетом оперативной памяти.
type MemLimiter interface {
	// Init ограничитель памяти инициализируется лениво из-за циклических связей с Servus
	Init(ss servus.Servus) error
	// MakeUnaryServerInterceptor возвращает интерсептор для унарных запросов
	MakeUnaryServerInterceptor() grpc.UnaryServerInterceptor
	// MakeStreamServerInterceptor возвращает интерсептор для стримовых запросов
	MakeStreamServerInterceptor() grpc.StreamServerInterceptor
	// Quit корректно завершает работу сервиса
	Quit()
}
