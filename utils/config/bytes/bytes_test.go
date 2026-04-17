/*
 * Copyright (c) New Cloud Technologies, Ltd. 2013-2022.
 * Author: Vitaly Isaev <vitaly.isaev@myoffice.team>
 * License: https://github.com/newcloudtechnologies/memlimiter/blob/master/LICENSE
 */

package bytes

import (
	"encoding/json"
	"testing"

	"code.cloudfoundry.org/bytefmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testStruct struct {
	Size Bytes `json:"size"`
}

func TestSize_UnmarshalJSON(t *testing.T) {
	var ts testStruct

	data := []byte(`{"size": "20M"}`)
	require.NoError(t, json.Unmarshal(data, &ts))
	assert.Equal(t, uint64(20*bytefmt.MEGABYTE), ts.Size.Value)

	data = []byte(`{"size":"invalid"}`)
	require.Error(t, json.Unmarshal(data, &ts))

	data = []byte(`{"size":"30MB"}`)
	require.NoError(t, json.Unmarshal(data, &ts))
	assert.Equal(t, uint64(30*bytefmt.MEGABYTE), ts.Size.Value)

	data = []byte(`{"size":"40K"}`)
	require.NoError(t, json.Unmarshal(data, &ts))
	assert.Equal(t, uint64(40*bytefmt.KILOBYTE), ts.Size.Value)

	data = []byte(`{"size":"50KB"}`)
	require.NoError(t, json.Unmarshal(data, &ts))
	assert.Equal(t, uint64(50*bytefmt.KILOBYTE), ts.Size.Value)

	// also check lowercase
	data = []byte(`{"size":"50kb"}`)
	require.NoError(t, json.Unmarshal(data, &ts))
	assert.Equal(t, uint64(50*bytefmt.KILOBYTE), ts.Size.Value)
}

func TestSize_MarshalJSON(t *testing.T) {
	var ts testStruct

	ts.Size = Bytes{Value: 20 * bytefmt.MEGABYTE}
	data, err := json.Marshal(&ts)
	require.NoError(t, err)
	assert.JSONEq(t, `{"size":"20M"}`, string(data))

	ts.Size = Bytes{Value: 40 * bytefmt.KILOBYTE}
	data, err = json.Marshal(&ts)
	require.NoError(t, err)
	assert.JSONEq(t, `{"size":"40K"}`, string(data))

	ts.Size = Bytes{Value: 1 * bytefmt.BYTE}
	data, err = json.Marshal(&ts)
	require.NoError(t, err)
	assert.JSONEq(t, `{"size":"1B"}`, string(data))
}

func TestBytesByValue(t *testing.T) {
	type masterStructVal struct {
		T testStruct `json:"t"`
	}

	var ms masterStructVal

	data := []byte(`{"t":{"size":"20M"}}`)
	require.NoError(t, json.Unmarshal(data, &ms))
	assert.Equal(t, uint64(20*bytefmt.MEGABYTE), ms.T.Size.Value)

	dump, err := json.Marshal(&ms)
	require.NoError(t, err)
	assert.JSONEq(t, `{"t":{"size":"20M"}}`, string(dump))
}

func TestBytesByPointer(t *testing.T) {
	type masterStructPtr struct {
		T *testStruct `json:"t"`
	}

	var ms masterStructPtr

	data := []byte(`{"t":{"size":"20M"}}`)
	require.NoError(t, json.Unmarshal(data, &ms))
	assert.Equal(t, uint64(20*bytefmt.MEGABYTE), ms.T.Size.Value)

	dump, err := json.Marshal(&ms)
	require.NoError(t, err)
	assert.JSONEq(t, `{"t":{"size":"20M"}}`, string(dump))
}

func TestBytesZeroValue(t *testing.T) {
	var ts testStruct

	data := []byte(`{"size": "20M"}`)
	require.NoError(t, json.Unmarshal(data, &ts))
	assert.Equal(t, uint64(20*bytefmt.MEGABYTE), ts.Size.Value)

	data = []byte(`{"size": "0"}`)
	require.NoError(t, json.Unmarshal(data, &ts))
	assert.Equal(t, uint64(0), ts.Size.Value)
}
