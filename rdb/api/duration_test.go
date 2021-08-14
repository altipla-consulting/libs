package api

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestMarshal(t *testing.T) {
	dur := Duration(3*time.Hour + 4*time.Minute + 5*time.Second)
	result, err := dur.MarshalJSON()
	require.NoError(t, err)

	require.Equal(t, string(result), `"03:04:05"`)
}

func TestMarshalDays(t *testing.T) {
	dur := Duration(2*24*time.Hour + 3*time.Hour + 4*time.Minute + 5*time.Second)
	result, err := dur.MarshalJSON()
	require.NoError(t, err)

	require.Equal(t, string(result), `"2.03:04:05"`)
}

func TestUnmarshal(t *testing.T) {
	dur := Duration(0)
	require.NoError(t, dur.UnmarshalJSON([]byte(`"03:04:05"`)))

	require.EqualValues(t, dur, 3*time.Hour+4*time.Minute+5*time.Second)
}

func TestUnmarshalDays(t *testing.T) {
	dur := Duration(0)
	require.NoError(t, dur.UnmarshalJSON([]byte(`"2.03:04:05"`)))

	require.EqualValues(t, dur, 2*24*time.Hour+3*time.Hour+4*time.Minute+5*time.Second)
}
