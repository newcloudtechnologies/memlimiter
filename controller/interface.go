package controller

import (
	"github.com/newcloudtechnologies/memlimiter/stats"
)

// Controller - generic memory consumption controller interface
type Controller interface {
	GetStats() (*stats.ControllerStats, error)
	Quit()
}
