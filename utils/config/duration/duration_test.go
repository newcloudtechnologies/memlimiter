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
	"github.com/stretchr/testify/require"
)

type testStruct struct {
	Timeout Duration `json:"timeout"`
}

func TestDuration_UnmarshalJSON(t *testing.T) {
	var ts testStruct

	data := []byte(`{ "timeout": "2ns" }`)
	require.NoError(t, json.Unmarshal(data, &ts))
	assert.Equal(t, 2*time.Nanosecond, ts.Timeout.Duration)

	data = []byte(`{ "timeout": "2ms" }`)
	require.NoError(t, json.Unmarshal(data, &ts))
	assert.Equal(t, 2*time.Millisecond, ts.Timeout.Duration)

	data = []byte(`{ "timeout": "2s" }`)
	require.NoError(t, json.Unmarshal(data, &ts))
	assert.Equal(t, 2*time.Second, ts.Timeout.Duration)

	data = []byte(`{ "timeout": "2m" }`)
	require.NoError(t, json.Unmarshal(data, &ts))
	assert.Equal(t, 2*time.Minute, ts.Timeout.Duration)

	data = []byte(`{ "timeout": "2h" }`)
	require.NoError(t, json.Unmarshal(data, &ts))
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
	require.NoError(t, err)
	assert.JSONEq(t, `{"timeout":"2ns"}`, string(dump))

	ts.Timeout = Duration{Duration: 2 * time.Millisecond}
	dump, err = json.Marshal(&ts)
	require.NoError(t, err)
	assert.JSONEq(t, `{"timeout":"2ms"}`, string(dump))

	ts.Timeout = Duration{Duration: 2 * time.Second}
	dump, err = json.Marshal(&ts)
	require.NoError(t, err)
	assert.JSONEq(t, `{"timeout":"2s"}`, string(dump))

	ts.Timeout = Duration{Duration: 2 * time.Minute}
	dump, err = json.Marshal(&ts)
	require.NoError(t, err)
	assert.JSONEq(t, `{"timeout":"2m0s"}`, string(dump))

	ts.Timeout = Duration{Duration: 2 * time.Hour}
	dump, err = json.Marshal(&ts)
	require.NoError(t, err)
	assert.JSONEq(t, `{"timeout":"2h0m0s"}`, string(dump))
}

func TestDurationByValue(t *testing.T) {
	type masterStructVal struct {
		T testStruct `json:"t"`
	}

	var ms masterStructVal

	data := []byte(`{"t":{"timeout":"2ns"}}`)
	require.NoError(t, json.Unmarshal(data, &ms))
	assert.Equal(t, 2*time.Nanosecond, ms.T.Timeout.Duration)

	dump, err := json.Marshal(&ms)
	require.NoError(t, err)
	assert.JSONEq(t, `{"t":{"timeout":"2ns"}}`, string(dump))
}

func TestDurationByPointer(t *testing.T) {
	type masterStructPtr struct {
		T *testStruct `json:"t"`
	}

	var ms masterStructPtr

	data := []byte(`{"t":{"timeout":"2ns"}}`)
	require.NoError(t, json.Unmarshal(data, &ms))
	assert.Equal(t, 2*time.Nanosecond, ms.T.Timeout.Duration)

	dump, err := json.Marshal(&ms)
	require.NoError(t, err)
	assert.JSONEq(t, `{"t":{"timeout":"2ns"}}`, string(dump))
}

func TestDurationZeroValue(t *testing.T) {
	var ts testStruct

	data := []byte(`{"timeout":"2s"}`)
	require.NoError(t, json.Unmarshal(data, &ts))
	assert.Equal(t, 2*time.Second, ts.Timeout.Duration)

	data = []byte(`{"timeout":"0"}`)
	require.NoError(t, json.Unmarshal(data, &ts))
	assert.Equal(t, 0*time.Second, ts.Timeout.Duration)

	data = []byte(`{"timeout":""}`)
	require.NoError(t, json.Unmarshal(data, &ts))
	assert.Equal(t, 0*time.Second, ts.Timeout.Duration)
}
