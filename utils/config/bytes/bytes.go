package bytes

import (
	"encoding/json"
	"fmt"

	"code.cloudfoundry.org/bytefmt"
)

// Bytes предоставляет возможность задавать конфигурируемые
// в json размеры в текстовом виде
type Bytes struct {
	Value uint64
}

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

func (b Bytes) MarshalJSON() ([]byte, error) {
	str := fmt.Sprintf("\"%s\"", bytefmt.ByteSize(b.Value))
	return []byte(str), nil
}

func (b Bytes) String() string { return bytefmt.ByteSize(b.Value) }
