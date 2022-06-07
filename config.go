package memlimiter

import (
	"github.com/pkg/errors"

	"github.com/newcloudtechnologies/memlimiter/controller/nextgc"
)

// Config - high-level MemLimiter config.
type Config struct {
	// ControllerNextGC - NextGC-based controller
	ControllerNextGC *nextgc.ControllerConfig `json:"controller_nextgc"` //nolint:tagliatelle
	// TODO:
	//  if new controller implementation appears, put its config here and make switch in Prepare()
	//  (only one subsection must be not nil).
}

// Prepare validates config.
func (c *Config) Prepare() error {
	if c.ControllerNextGC == nil {
		return errors.New("empty ControllerNextGC")
	}

	return nil
}
