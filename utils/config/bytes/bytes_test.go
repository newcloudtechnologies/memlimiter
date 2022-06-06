package bytes

import (
	"encoding/json"
	"testing"

	"code.cloudfoundry.org/bytefmt"
	"github.com/stretchr/testify/assert"
)

type testStruct struct {
	Size Bytes `json:"size"`
}

func TestSize_UnmarshalJSON(t *testing.T) {
	var ts testStruct

	data := []byte(`{"size": "20M"}`)
	assert.NoError(t, json.Unmarshal(data, &ts))
	assert.Equal(t, uint64(20*bytefmt.MEGABYTE), ts.Size.Value)

	data = []byte(`{"size":"invalid"}`)
	assert.Error(t, json.Unmarshal(data, &ts))

	data = []byte(`{"size":"30MB"}`)
	assert.NoError(t, json.Unmarshal(data, &ts))
	assert.Equal(t, uint64(30*bytefmt.MEGABYTE), ts.Size.Value)

	data = []byte(`{"size":"40K"}`)
	assert.NoError(t, json.Unmarshal(data, &ts))
	assert.Equal(t, uint64(40*bytefmt.KILOBYTE), ts.Size.Value)

	data = []byte(`{"size":"50KB"}`)
	assert.NoError(t, json.Unmarshal(data, &ts))
	assert.Equal(t, uint64(50*bytefmt.KILOBYTE), ts.Size.Value)

	// also check lowercase
	data = []byte(`{"size":"50kb"}`)
	assert.NoError(t, json.Unmarshal(data, &ts))
	assert.Equal(t, uint64(50*bytefmt.KILOBYTE), ts.Size.Value)
}

func TestSize_MarshalJSON(t *testing.T) {
	var ts testStruct

	ts.Size = Bytes{Value: 20 * bytefmt.MEGABYTE}
	data, err := json.Marshal(&ts)
	assert.NoError(t, err)
	assert.Equal(t, []byte(`{"size":"20M"}`), data)

	ts.Size = Bytes{Value: 40 * bytefmt.KILOBYTE}
	data, err = json.Marshal(&ts)
	assert.NoError(t, err)
	assert.Equal(t, []byte(`{"size":"40K"}`), data)

	ts.Size = Bytes{Value: 1 * bytefmt.BYTE}
	data, err = json.Marshal(&ts)
	assert.NoError(t, err)
	assert.Equal(t, []byte(`{"size":"1B"}`), data)
}

func TestBytesByValue(t *testing.T) {
	type masterStructVal struct {
		T testStruct `json:"t"`
	}

	var ms masterStructVal
	data := []byte(`{"t":{"size":"20M"}}`)
	assert.NoError(t, json.Unmarshal(data, &ms))
	assert.Equal(t, uint64(20*bytefmt.MEGABYTE), ms.T.Size.Value)

	dump, err := json.Marshal(&ms)
	assert.NoError(t, err)
	assert.Equal(t, []byte(`{"t":{"size":"20M"}}`), dump)
}

func TestBytesByPointer(t *testing.T) {
	type masterStructPtr struct {
		T *testStruct `json:"t"`
	}

	var ms masterStructPtr
	data := []byte(`{"t":{"size":"20M"}}`)
	assert.NoError(t, json.Unmarshal(data, &ms))
	assert.Equal(t, uint64(20*bytefmt.MEGABYTE), ms.T.Size.Value)

	dump, err := json.Marshal(&ms)
	assert.NoError(t, err)
	assert.Equal(t, []byte(`{"t":{"size":"20M"}}`), dump)
}

func TestBytesZeroValue(t *testing.T) {
	var ts testStruct
	data := []byte(`{"size": "0"}`)
	assert.NoError(t, json.Unmarshal(data, &ts))
	assert.Equal(t, uint64(0*bytefmt.MEGABYTE), ts.Size.Value)
}
