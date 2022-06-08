/*
 * Copyright (c) New Cloud Technologies, Ltd. 2013-2022.
 * Author: Vitaly Isaev <vitaly.isaev@myoffice.team>
 * License: https://github.com/newcloudtechnologies/memlimiter/blob/master/LICENSE
 */

package duration

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type testStruct struct {
	Timeout Duration `json:"timeout"`
}

func TestDuration_UnmarshalJSON(t *testing.T) {
	var ts testStruct

	data := []byte(`{ "timeout": "2ns" }`)
	assert.NoError(t, json.Unmarshal(data, &ts))
	assert.Equal(t, 2*time.Nanosecond, ts.Timeout.Duration)

	data = []byte(`{ "timeout": "2ms" }`)
	assert.NoError(t, json.Unmarshal(data, &ts))
	assert.Equal(t, 2*time.Millisecond, ts.Timeout.Duration)

	data = []byte(`{ "timeout": "2s" }`)
	assert.NoError(t, json.Unmarshal(data, &ts))
	assert.Equal(t, 2*time.Second, ts.Timeout.Duration)

	data = []byte(`{ "timeout": "2m" }`)
	assert.NoError(t, json.Unmarshal(data, &ts))
	assert.Equal(t, 2*time.Minute, ts.Timeout.Duration)

	data = []byte(`{ "timeout": "2h" }`)
	assert.NoError(t, json.Unmarshal(data, &ts))
	assert.Equal(t, 2*time.Hour, ts.Timeout.Duration)

	data = []byte(`{ "timeout": "invalid" }`)
	assert.Error(t, json.Unmarshal(data, &ts))
}

func TestDuration_MarshalJSON(t *testing.T) {
	var (
		ts   testStruct
		dump []byte
		err  error
	)

	ts.Timeout = Duration{Duration: 2 * time.Nanosecond}
	dump, err = json.Marshal(&ts)
	assert.NoError(t, err)
	assert.Equal(t, []byte(`{"timeout":"2ns"}`), dump)

	ts.Timeout = Duration{Duration: 2 * time.Millisecond}
	dump, err = json.Marshal(&ts)
	assert.NoError(t, err)
	assert.Equal(t, []byte(`{"timeout":"2ms"}`), dump)

	ts.Timeout = Duration{Duration: 2 * time.Second}
	dump, err = json.Marshal(&ts)
	assert.NoError(t, err)
	assert.Equal(t, []byte(`{"timeout":"2s"}`), dump)

	ts.Timeout = Duration{Duration: 2 * time.Minute}
	dump, err = json.Marshal(&ts)
	assert.NoError(t, err)
	assert.Equal(t, []byte(`{"timeout":"2m0s"}`), dump)

	ts.Timeout = Duration{Duration: 2 * time.Hour}
	dump, err = json.Marshal(&ts)
	assert.NoError(t, err)
	assert.Equal(t, []byte(`{"timeout":"2h0m0s"}`), dump)
}

func TestDurationByValue(t *testing.T) {
	type masterStructVal struct {
		T testStruct `json:"t"`
	}

	var ms masterStructVal

	data := []byte(`{"t":{"timeout":"2ns"}}`)
	assert.NoError(t, json.Unmarshal(data, &ms))
	assert.Equal(t, 2*time.Nanosecond, ms.T.Timeout.Duration)

	dump, err := json.Marshal(&ms)
	assert.NoError(t, err)
	assert.Equal(t, []byte(`{"t":{"timeout":"2ns"}}`), dump)
}

func TestDurationByPointer(t *testing.T) {
	type masterStructPtr struct {
		T *testStruct `json:"t"`
	}

	var ms masterStructPtr

	data := []byte(`{"t":{"timeout":"2ns"}}`)
	assert.NoError(t, json.Unmarshal(data, &ms))
	assert.Equal(t, 2*time.Nanosecond, ms.T.Timeout.Duration)

	dump, err := json.Marshal(&ms)
	assert.NoError(t, err)
	assert.Equal(t, []byte(`{"t":{"timeout":"2ns"}}`), dump)
}

func TestDurationZeroValue(t *testing.T) {
	var ts testStruct

	data := []byte(`{"size": "0"}`)
	assert.NoError(t, json.Unmarshal(data, &ts))
	assert.Equal(t, 0*time.Second, ts.Timeout.Duration)
}
