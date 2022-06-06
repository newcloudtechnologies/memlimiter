package stats

type Service struct {
	NextGC uint64
	Custom interface{}
}

// Subscription - интерфейс подписки на оперативную статистику
type Subscription interface {
	Updates() <-chan *Service
	Quit()
}
