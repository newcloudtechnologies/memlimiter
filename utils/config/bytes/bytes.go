package bytes

import (
	"encoding/json"
	"fmt"

	"code.cloudfoundry.org/bytefmt"
)

// Bytes helps to represent human-readable size values in JSON.
type Bytes struct {
	Value uint64
}

// UnmarshalJSON - JSON deserializer.
func (b *Bytes) UnmarshalJSON(data []byte) (err error) {
	var s string

	if err = json.Unmarshal(data, &s); err != nil {
		return
	}

	if s == "0" {
		return
	}

	b.Value, err = bytefmt.ToBytes(s)

	return
}

// MarshalJSON - JSON serializer.
func (b Bytes) MarshalJSON() ([]byte, error) {
	str := fmt.Sprintf("\"%s\"", bytefmt.ByteSize(b.Value))

	return []byte(str), nil
}

func (b Bytes) String() string { return bytefmt.ByteSize(b.Value) }
