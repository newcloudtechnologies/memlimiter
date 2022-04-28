package memlimiter

import (
	"gitlab.stageoffice.ru/UCS-COMMON/gaben"

	"github.com/newcloudtechnologies/memlimiter/utils"
)

// NewMemLimiterFromConfig - конструктор системы управления бюджетом памяти.
func NewMemLimiterFromConfig(
	logger gaben.Logger,
	cfg *Config,
	applicationTerminator utils.ApplicationTerminator,
	consumptionReporter utils.ConsumptionReporter,
) (MemLimiter, error) {
	if cfg == nil {
		// передача nil конфига означает, что MemLimiter отключён, и вместо него будет заглушка
		return &memLimiterStub{}, nil
	}

	return newMemLimiterDefault(logger, cfg, applicationTerminator, consumptionReporter)
}
