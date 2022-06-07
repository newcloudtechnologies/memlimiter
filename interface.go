package memlimiter

import (
	"github.com/newcloudtechnologies/memlimiter/stats"
	"google.golang.org/grpc"
)

// Service - верхнеуровневый интерфейс системы управления бюджетом оперативной памяти.
type Service interface {
	GetStats() (*stats.MemLimiterStats, error)
	// MakeUnaryServerInterceptor возвращает интерсептор для унарных запросов
	MakeUnaryServerInterceptor() grpc.UnaryServerInterceptor
	// MakeStreamServerInterceptor возвращает интерсептор для стримовых запросов
	MakeStreamServerInterceptor() grpc.StreamServerInterceptor
	// Quit корректно завершает работу сервиса
	Quit()
}
