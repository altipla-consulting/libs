package rdb

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNestedOr(t *testing.T) {
	q := And(Filter("foo", 1), Or(Filter("bar", 2), Filter("baz", 3)), Filter("qux", 4))
	params := NewParams()
	rql := q.RQL(params)

	require.Equal(t, rql, `foo = $p0 and (bar = $p1 or baz = $p2) and qux = $p3`)
	require.Equal(t, params.values, map[string]interface{}{
		"p0": 1,
		"p1": 2,
		"p2": 3,
		"p3": 4,
	})
}

func TestParamsTime(t *testing.T) {
	p := NewParams()
	require.Equal(t, p.Next(time.Date(2019, time.January, 2, 3, 4, 5, 0, time.UTC)), "$p0")
	require.Equal(t, p.values["p0"], "2019-01-02T03:04:05.0000000Z")
}
