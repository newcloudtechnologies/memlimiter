package memlimiter

import (
	"github.com/go-logr/logr"
	"github.com/newcloudtechnologies/memlimiter/utils"
)

// NewMemLimiterFromConfig - конструктор системы управления бюджетом памяти.
func NewMemLimiterFromConfig(
	logger logr.Logger,
	cfg *Config,
	applicationTerminator utils.ApplicationTerminator,
	consumptionReporter utils.ConsumptionReporter,
) (MemLimiter, error) {
	return newMemLimiterDefault(logger, cfg, applicationTerminator, consumptionReporter)
}
