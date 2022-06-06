package duration

import (
	"encoding/json"
	"fmt"
	"time"
)

// Duration предоставляет возможность задавать конфигурируемые
// в json длительности в текстовом виде
type Duration struct {
	time.Duration
}

// UnmarshalJSON ...
func (d *Duration) UnmarshalJSON(data []byte) (err error) {
	var s string

	if err = json.Unmarshal(data, &s); err != nil {
		return
	}

	if s == "0" { // для случая без указания размерности
		return
	}
	if s == "" { // для случая использования в cli-интерфейсе и пустой строки в качестве дефолта
		return
	}

	d.Duration, err = time.ParseDuration(s)
	return
}

// MarshalJSON ...
func (d Duration) MarshalJSON() ([]byte, error) {
	s := fmt.Sprintf("\"%s\"", d.Duration.String())
	return []byte(s), nil
}
