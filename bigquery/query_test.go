package bigquery

import (
	"testing"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/stretchr/testify/require"
)

var exampleTable = Table("example")

func buildSQL(query *Query) *bigquery.Query {
	b := new(sqlBuilder)
	q := new(bigquery.Client).Query(query.buildSQL("exampleds", b))
	q.Parameters = b.params
	return q
}

func TestQuerySimplest(t *testing.T) {
	q := NewQuery(exampleTable, Select("foo", "bar"))
	req := buildSQL(q)

	require.Equal(t, req.Q, "SELECT foo, bar FROM exampleds.example")
	require.Empty(t, req.Parameters)
}

func TestQueryPartition(t *testing.T) {
	start := time.Date(2019, 10, 13, 0, 0, 0, 0, time.UTC)
	end := time.Date(2019, 10, 15, 0, 0, 0, 0, time.UTC)
	q := NewQuery(exampleTable, Select("foo"), WithPartitionRange(start, end))
	req := buildSQL(q)

	require.Equal(t, req.Q, `SELECT foo FROM exampleds.example WHERE _PARTITIONTIME BETWEEN "2019-10-13" AND "2019-10-16"`)
	require.Empty(t, req.Parameters)
}

func TestQueryOrderAsc(t *testing.T) {
	q := NewQuery(exampleTable, Select("foo"))
	q = q.OrderBy("bar")
	req := buildSQL(q)

	require.Equal(t, req.Q, "SELECT foo FROM exampleds.example ORDER BY bar ASC")
	require.Empty(t, req.Parameters)
}

func TestQueryOrderDesc(t *testing.T) {
	q := NewQuery(exampleTable, Select("foo"))
	q = q.OrderBy("-bar")
	req := buildSQL(q)

	require.Equal(t, req.Q, "SELECT foo FROM exampleds.example ORDER BY bar DESC")
	require.Empty(t, req.Parameters)
}

func TestQueryLimit(t *testing.T) {
	q := NewQuery(exampleTable, Select("foo"))
	q = q.Limit(500)
	req := buildSQL(q)

	require.Equal(t, req.Q, "SELECT foo FROM exampleds.example LIMIT 500")
	require.Empty(t, req.Parameters)
}

func TestQuerySingleFilter(t *testing.T) {
	q := NewQuery(exampleTable, Select("foo"))
	q = q.Filter("foo", 3)
	req := buildSQL(q)

	require.Equal(t, req.Q, "SELECT foo FROM exampleds.example WHERE foo = @p0")
	require.Len(t, req.Parameters, 1)
	require.Equal(t, req.Parameters[0].Name, "p0")
	require.EqualValues(t, req.Parameters[0].Value, 3)
}

func TestQueryMultipleFilters(t *testing.T) {
	q := NewQuery(exampleTable, Select("foo"))
	q = q.Filter("foo", 3)
	q = q.Filter("bar", "mystring")
	req := buildSQL(q)

	require.Equal(t, req.Q, "SELECT foo FROM exampleds.example WHERE foo = @p0 AND bar = @p1")
	require.Len(t, req.Parameters, 2)
	require.Equal(t, req.Parameters[0].Name, "p0")
	require.EqualValues(t, req.Parameters[0].Value, 3)
	require.Equal(t, req.Parameters[1].Name, "p1")
	require.Equal(t, req.Parameters[1].Value, "mystring")
}

func TestQueryMultipleFiltersPartition(t *testing.T) {
	start := time.Date(2019, 10, 13, 0, 0, 0, 0, time.UTC)
	end := time.Date(2019, 10, 15, 0, 0, 0, 0, time.UTC)
	q := NewQuery(exampleTable, Select("foo", "bar"), WithPartitionRange(start, end))
	q = q.Filter("foo", 3)
	q = q.Filter("bar", "mystring")
	req := buildSQL(q)

	require.Equal(t, req.Q, `SELECT foo, bar FROM exampleds.example WHERE _PARTITIONTIME BETWEEN "2019-10-13" AND "2019-10-16" AND foo = @p0 AND bar = @p1`)
	require.Len(t, req.Parameters, 2)
	require.Equal(t, req.Parameters[0].Name, "p0")
	require.EqualValues(t, req.Parameters[0].Value, 3)
	require.Equal(t, req.Parameters[1].Name, "p1")
	require.Equal(t, req.Parameters[1].Value, "mystring")
}

func TestQueryFilterOr(t *testing.T) {
	q := NewQuery(exampleTable, Select("foo"))
	cond := ConditionOr()
	cond = cond.Filter("foo", 3)
	cond = cond.Filter("bar", 4)
	q = q.FilterCond(cond)
	q = q.Filter("baz", 5)
	req := buildSQL(q)

	require.Equal(t, req.Q, "SELECT foo FROM exampleds.example WHERE baz = @p0 AND (foo = @p1 OR bar = @p2)")
	require.Len(t, req.Parameters, 3)
	require.Equal(t, req.Parameters[0].Name, "p0")
	require.EqualValues(t, req.Parameters[0].Value, 5)
	require.Equal(t, req.Parameters[1].Name, "p1")
	require.EqualValues(t, req.Parameters[1].Value, 3)
	require.Equal(t, req.Parameters[2].Name, "p2")
	require.EqualValues(t, req.Parameters[2].Value, 4)
}

func TestQueryChildConditionEmpty(t *testing.T) {
	q := NewQuery(exampleTable, Select("foo"))
	cond := ConditionOr()
	q = q.FilterCond(cond)
	req := buildSQL(q)

	require.Equal(t, req.Q, "SELECT foo FROM exampleds.example")
	require.Empty(t, req.Parameters)
}
