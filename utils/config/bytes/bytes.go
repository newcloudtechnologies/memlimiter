/*
 * Copyright (c) New Cloud Technologies, Ltd. 2013-2026.
 * Author: Vitaly Isaev <vitaly.isaev@myoffice.team>
 * License: https://github.com/newcloudtechnologies/memlimiter/blob/master/LICENSE
 */

package bytes

import (
	"encoding/json"

	"code.cloudfoundry.org/bytefmt"
)

// Bytes helps to represent human-readable size values in JSON.
type Bytes struct {
	// Value is the number of bytes.
	Value uint64
}

// UnmarshalJSON parses a JSON string like "512M" into bytes.
func (b *Bytes) UnmarshalJSON(data []byte) error {
	var s string

	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	if s == "0" {
		b.Value = 0

		return nil
	}

	value, err := bytefmt.ToBytes(s)
	if err != nil {
		return err
	}

	b.Value = value

	return nil
}

// MarshalJSON renders bytes as a human-readable JSON string (for example, "20M").
func (b Bytes) MarshalJSON() ([]byte, error) {
	return json.Marshal(bytefmt.ByteSize(b.Value))
}
