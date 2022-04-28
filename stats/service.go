package stats

type Service struct {
	NextGC uint64
}

// ServiceSubscription - интерфейс подписки на оперативную статистику
type ServiceSubscription interface {
	Updates() <-chan *Service
	Quit()
}
