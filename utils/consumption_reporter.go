package utils

// ConsumptionReporter предоставляет информацию о заранее известных крупных структурах - потребителях памяти в составе процесса
// (это могут быть кеши, пулы памяти и другие объекты).
type ConsumptionReporter interface {
	// PredefinedConsumers возвращает отчёт о потреблении памяти специализированными структурами.
	// Исходные данные для построения отчёта могут быть извлечены из агрегированной статистики сервиса (передаётся первым аргументом).
	// Но сервис не обязан опираться на неё, если у него есть свои источники информации.
	PredefinedConsumers(serviceStats interface{}) (*ConsumptionReport, error)
}

// ConsumptionReport - отчёт о расходах памяти крупными структурами данных в составе процесса (кешами, пулами памяти и др.)
type ConsumptionReport struct {
	// Go - показатели потребления памяти структурами, память которым выделяется с помощью стандартного аллокатора Go.
	// (ключ - произвольная строка, значение - байты).
	Go map[string]uint64
	// Cgo - показатели потребления памяти структурами, память которым выделяется другими аллокаторами за границей Cgo.
	// (ключ - произвольная строка, значение - байты).
	Cgo map[string]uint64
}
