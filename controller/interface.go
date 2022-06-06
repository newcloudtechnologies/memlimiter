package controller

import (
	"github.com/newcloudtechnologies/memlimiter/stats"
)

// Controller - обобщённый интерфейс регулятора.
type Controller interface {
	GetStats() (*stats.Controller, error)
	Quit()
}
