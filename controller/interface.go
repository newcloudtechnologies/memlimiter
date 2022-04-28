package controller

import (
	servus_stats "gitlab.stageoffice.ru/UCS-COMMON/schemagen-go/v41/servus/stats/v1"
)

// Controller - обобщённый интерфейс регулятора.
type Controller interface {
	GetStats() (*servus_stats.GoMemLimiterStats_ControllerStats, error)
	Quit()
}
