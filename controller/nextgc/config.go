package nextgc

import (
	"github.com/newcloudtechnologies/memlimiter/utils/config/bytes"
	"github.com/newcloudtechnologies/memlimiter/utils/config/duration"
	"github.com/pkg/errors"
)

// ControllerConfig - controller configuration.
type ControllerConfig struct {
	// RSSLimit - physical memory (RSS) consumption hard limit for a process.
	RSSLimit bytes.Bytes `json:"rss_limit"`
	// DangerZoneGOGC - RSS utilization threshold that triggers controller to
	// set more conservative parameters for GC.
	// Possible values are in range (0; 100).
	DangerZoneGOGC uint32 `json:"danger_zone_gogc"`
	// DangerZoneGOGC - RSS utilization threshold that triggers controller to
	// throttle incoming requests.
	// Possible values are in range (0; 100).
	DangerZoneThrottling uint32 `json:"danger_zone_throttling"`
	// Period - the periodicity of control parameters computation.
	Period duration.Duration `json:"period"`
	// ComponentProportional - controller's proportional component configuration
	ComponentProportional *ComponentProportionalConfig `json:"component_proportional"`
	// TODO:
	//   if some other components will appear in future, put their configs here.
}

// Prepare - config validator.
func (c *ControllerConfig) Prepare() error {
	if c.RSSLimit.Value == 0 {
		return errors.New("empty RSSLimit")
	}

	if c.DangerZoneGOGC == 0 || c.DangerZoneGOGC > 100 {
		return errors.New("invalid DangerZoneGOGC value (must belong to [0; 100])")
	}

	if c.DangerZoneThrottling == 0 || c.DangerZoneThrottling > 100 {
		return errors.Errorf("invalid DangerZoneThrottling value (must belong to [0; 100])")
	}

	if c.Period.Duration == 0 {
		return errors.New("empty Period")
	}

	if c.ComponentProportional == nil {
		return errors.New("empty ComponentProportional")
	}

	return nil
}

// ComponentProportionalConfig - controller's proportional component configuration.
type ComponentProportionalConfig struct {
	// Coefficient - coefficient used to computed weighted sum of in the controller equation
	Coefficient float64 `json:"coefficient"`
	// WindowSize - averaging window size for the EMA. Averaging is disabled if WindowSize is zero.
	WindowSize uint `json:"window_size"`
}

// Prepare - config validator.
func (c *ComponentProportionalConfig) Prepare() error {
	if c.Coefficient == 0 {
		return errors.New("empty Coefficient makes no sense")
	}

	return nil
}
