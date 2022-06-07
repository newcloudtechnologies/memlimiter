package memlimiter

import (
	"github.com/newcloudtechnologies/memlimiter/middleware"
	"github.com/newcloudtechnologies/memlimiter/stats"
)

// Service - a high-level interface for a memory usage control subsystem.
type Service interface {
	Middleware() middleware.Middleware
	GetStats() (*stats.MemLimiterStats, error)
	// Quit terminates service gracefully.
	Quit()
}
