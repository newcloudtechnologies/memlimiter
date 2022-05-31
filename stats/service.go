package stats

type Service struct {
	NextGC uint64
}

type Subscription interface {
	Updates() <-chan *Service
	Quit()
}
