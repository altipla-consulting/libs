package encoding

import (
	"bytes"
	"encoding/gob"
	"testing"

	"github.com/stretchr/testify/require"
)

type normalStruct struct {
	Value *int32
}

func TestNormalEncodingNil(t *testing.T) {
	var buf bytes.Buffer

	value := int32(0)
	s := &normalStruct{
		Value: &value,
	}
	require.NoError(t, gob.NewEncoder(&buf).Encode(s))

	other := new(normalStruct)
	require.NoError(t, gob.NewDecoder(&buf).Decode(other))

	require.Nil(t, other.Value)
}

func TestNormalEncodingValue(t *testing.T) {
	var buf bytes.Buffer

	value := int32(5)
	s := &normalStruct{
		Value: &value,
	}
	require.NoError(t, gob.NewEncoder(&buf).Encode(s))

	other := new(normalStruct)
	require.NoError(t, gob.NewDecoder(&buf).Decode(other))

	require.NotNil(t, other.Value)
	require.EqualValues(t, *other.Value, 5)
}

type int32Struct struct {
	Value *GobInt32
}

func TestInt32EncodingZero(t *testing.T) {
	var buf bytes.Buffer

	s := &int32Struct{
		Value: NewGobInt32(0),
	}
	require.NoError(t, gob.NewEncoder(&buf).Encode(s))

	other := new(int32Struct)
	require.NoError(t, gob.NewDecoder(&buf).Decode(other))

	require.NotNil(t, other.Value)
	require.EqualValues(t, *other.Value, 0)
}

func TestInt32EncodingValue(t *testing.T) {
	var buf bytes.Buffer

	s := &int32Struct{
		Value: NewGobInt32(5),
	}
	require.NoError(t, gob.NewEncoder(&buf).Encode(s))

	other := new(int32Struct)
	require.NoError(t, gob.NewDecoder(&buf).Decode(other))

	require.NotNil(t, other.Value)
	require.EqualValues(t, *other.Value, 5)
}

func TestInt32EncodingNil(t *testing.T) {
	var buf bytes.Buffer

	s := new(int32Struct)
	require.NoError(t, gob.NewEncoder(&buf).Encode(s))

	other := new(int32Struct)
	require.NoError(t, gob.NewDecoder(&buf).Decode(other))

	require.Nil(t, other.Value)
}
