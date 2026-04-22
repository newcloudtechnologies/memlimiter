/*
 * Copyright (c) New Cloud Technologies, Ltd. 2013-2026.
 * Author: Vitaly Isaev <vitaly.isaev@myoffice.team>
 * License: https://github.com/newcloudtechnologies/memlimiter/blob/master/LICENSE
 */

package nextgc

import (
	"errors"

	"github.com/newcloudtechnologies/memlimiter/backpressure"
	"github.com/newcloudtechnologies/memlimiter/utils/config/bytes"
	"github.com/newcloudtechnologies/memlimiter/utils/config/duration"
)

// defaultMinGOGC is the default minimum GOGC value used in the "red zone".
const defaultMinGOGC = 10

// ControllerConfig - controller configuration.
type ControllerConfig struct {
	// RSSLimit - physical memory (RSS) consumption hard limit for a process.
	RSSLimit bytes.Bytes `json:"rss_limit"`
	// DangerZoneGOGC - RSS utilization threshold that triggers controller to
	// set more conservative parameters for GC.
	// Possible values are in range (0; 100).
	DangerZoneGOGC uint32 `json:"danger_zone_gogc"`
	// DangerZoneThrottling - RSS utilization threshold that triggers controller to
	// throttle incoming requests.
	// Possible values are in range (0; 100).
	// It's recommended to keep it greater than or equal to DangerZoneGOGC so that
	// the service first intensifies GC and starts throttling only later.
	DangerZoneThrottling uint32 `json:"danger_zone_throttling"`
	// Period - the periodicity of control parameters computation.
	Period duration.Duration `json:"period"`
	// MinGOGC - minimal allowed GOGC value used in the "red zone".
	// Zero means default safe value.
	MinGOGC int `json:"min_gogc"`
	// ComponentProportional - controller's proportional component configuration
	ComponentProportional *ComponentProportionalConfig `json:"component_proportional"`
	// TODO:
	//   if some other components will appear in future, put their configs here.
}

// Prepare - config validator.
func (c *ControllerConfig) Prepare() error {
	if err := c.validateRSSLimit(); err != nil {
		return err
	}

	if err := c.validateDangerZoneGOGC(); err != nil {
		return err
	}

	if err := c.validateDangerZoneThrottling(); err != nil {
		return err
	}

	if err := c.validatePeriod(); err != nil {
		return err
	}

	c.applyDefaults()

	if err := c.validateMinGOGC(); err != nil {
		return err
	}

	if err := c.validateComponentProportional(); err != nil {
		return err
	}

	return nil
}

func (c *ControllerConfig) validateRSSLimit() error {
	if c.RSSLimit.Value == 0 {
		return errors.New("empty RSSLimit")
	}

	return nil
}

func (c *ControllerConfig) validateDangerZoneGOGC() error {
	if c.DangerZoneGOGC == 0 || c.DangerZoneGOGC >= 100 {
		return errors.New("invalid DangerZoneGOGC value (must belong to (0; 100))")
	}

	return nil
}

func (c *ControllerConfig) validateDangerZoneThrottling() error {
	if c.DangerZoneThrottling == 0 || c.DangerZoneThrottling >= 100 {
		return errors.New("invalid DangerZoneThrottling value (must belong to (0; 100))")
	}

	return nil
}

func (c *ControllerConfig) validatePeriod() error {
	if c.Period.Duration == 0 {
		return errors.New("empty Period")
	}

	return nil
}

func (c *ControllerConfig) applyDefaults() {
	if c.MinGOGC == 0 {
		c.MinGOGC = defaultMinGOGC
	}
}

func (c *ControllerConfig) validateMinGOGC() error {
	if c.MinGOGC < 1 || c.MinGOGC > backpressure.DefaultGOGC {
		return errors.New("invalid MinGOGC value")
	}

	return nil
}

func (c *ControllerConfig) validateComponentProportional() error {
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
