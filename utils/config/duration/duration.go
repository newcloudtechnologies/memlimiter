/*
 * Copyright (c) New Cloud Technologies, Ltd. 2013-2022.
 * Author: Vitaly Isaev <vitaly.isaev@myoffice.team>
 * License: https://github.com/newcloudtechnologies/memlimiter/blob/master/LICENSE
 */

package duration

import (
	"encoding/json"
	"time"
)

// Duration helps to represent human-readable duration values in JSON.
type Duration struct {
	// Duration is the duration value.
	time.Duration
}

// UnmarshalJSON parses a JSON string like "500ms" into a duration value.
func (d *Duration) UnmarshalJSON(data []byte) error {
	var s string

	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	// Both "0" and empty string mean zero duration.
	if s == "0" || s == "" {
		d.Duration = 0

		return nil
	}

	value, err := time.ParseDuration(s)
	if err != nil {
		return err
	}

	d.Duration = value

	return nil
}

// MarshalJSON renders duration as a human-readable JSON string (for example, "2s").
func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.Duration.String())
}
