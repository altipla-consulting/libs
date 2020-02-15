package bigquery

import (
	"fmt"
	"hash/crc32"
	"strings"
	"time"

	"cloud.google.com/go/bigquery"
)

// Query is a SQL builder for BigQuery queries.
type Query struct {
	table    Table
	cols     []string
	orders   []string
	limit    int64
	rootCond *Condition
}

// QueryOption are options we can configure in a query.
type QueryOption func(q *Query)

// WithPartitionRange controls the range of partitions we want to query in a
// table to save money not processing unnecessary data.
//
// The range is according to API standards, that is: [start, end)
func WithPartitionRange(start, end time.Time) QueryOption {
	return func(q *Query) {
		q.rootCond = q.rootCond.Filter(fmt.Sprintf(`_PARTITIONTIME BETWEEN "%s" AND "%s"`, start.Format("2006-01-02"), end.AddDate(0, 0, 1).Format("2006-01-02")))
	}
}

// Select the columns to return in the query. This is a required option in all queries
// because there is no sane default to not consume BigQuery quota.
func Select(cols ...string) QueryOption {
	return func(q *Query) {
		q.cols = cols
	}
}

// NewQuery builds a new query to the table.
//
// There is a required option for all tables: Select(...)
// For partitioned tables there is another required option: WithPartitionRange(...)
func NewQuery(table Table, opts ...QueryOption) *Query {
	q := &Query{
		table:    table,
		rootCond: ConditionAnd(),
	}
	for _, opt := range opts {
		opt(q)
	}
	return q
}

// Clone returns a copy of the query.
func (q *Query) Clone() *Query {
	return &Query{
		table:    q.table,
		cols:     q.cols,
		orders:   q.orders,
		limit:    q.limit,
		rootCond: q.rootCond.Clone(),
	}
}

// OrderBy sets the order the results will be returned.
func (q *Query) OrderBy(column string) *Query {
	q = q.Clone()

	if strings.Contains(column, ",") {
		panic("call Order multiple times, do not pass multiple columns")
	}
	if strings.Contains(column, "ASC") {
		panic("do not call Order with `foo ASC`, use plain `foo` instead")
	}
	if strings.Contains(column, "DESC") {
		panic("do not call Order with `foo DESC`, use plain `-foo` instead")
	}

	if strings.HasPrefix(column, "-") {
		q.orders = append(q.orders, column[1:]+" DESC")
	} else {
		q.orders = append(q.orders, column+" ASC")
	}

	return q
}

// Limit sets the limit of results to be returned.
func (q *Query) Limit(limit int64) *Query {
	q = q.Clone()
	q.limit = limit
	return q
}

// Condition is a filter condition of a query.
type Condition struct {
	operator string
	filters  []sqlFilter
	children []*Condition
}

type sqlFilter struct {
	filter string
	values []interface{}
}

// ConditionAnd builds a new condition where all the members are merged with the AND operator.
func ConditionAnd() *Condition {
	return &Condition{
		operator: "AND",
	}
}

// ConditionAnd builds a new condition where all the members are merged with the OR operator.
func ConditionOr() *Condition {
	return &Condition{
		operator: "OR",
	}
}

// Clone returns a copy of the condition.
func (cond *Condition) Clone() *Condition {
	return &Condition{
		operator: cond.operator,
		filters:  cond.filters,
		children: cond.children,
	}
}

// Filter builds a new condition adding the filter we specify. Examples:
//
//   .Filter("foo", 3)
//   .Filter("foo >", 3)
//   .Filter("foo BETWEEN ? AND ?", 3, 4)
func (cond *Condition) Filter(filter string, values ...interface{}) *Condition {
	if len(values) == 1 {
		if !strings.Contains(filter, " ") {
			filter = filter + " = ?"
		} else if !strings.Contains(filter, "?") {
			filter = filter + " ?"
		}
	}

	cond = cond.Clone()
	cond.filters = append(cond.filters, sqlFilter{
		filter: filter,
		values: values,
	})
	return cond
}

// FilterCond builds a new condition adding the new child condition we specify here. Examples:
//
//   cond := ConditionOr()
//   cond = cond.Filter("foo", 3)
//   cond = cond.Filter("bar", 4)
//   .FilterCond(cond)
func (cond *Condition) FilterCond(child *Condition) *Condition {
	cond = cond.Clone()
	cond.children = append(cond.children, child)
	return cond
}

func (c *Condition) buildSQL(b *sqlBuilder) string {
	if len(c.filters) == 0 && len(c.children) == 0 {
		return ""
	}

	sql := make([]string, 0, len(c.filters)+len(c.children))
	for _, f := range c.filters {
		for _, value := range f.values {
			f.filter = strings.Replace(f.filter, "?", b.addParam(value), 1)
		}
		sql = append(sql, f.filter)
	}

	for _, child := range c.children {
		// Do not add parentheses for children without conditions.
		if len(child.filters) == 0 && len(child.children) == 0 {
			continue
		}

		sql = append(sql, "("+child.buildSQL(b)+")")
	}

	return strings.Join(sql, " "+c.operator+" ")
}

type sqlBuilder struct {
	params []bigquery.QueryParameter
}

func (b *sqlBuilder) addParam(value interface{}) string {
	name := fmt.Sprintf("p%d", len(b.params))
	b.params = append(b.params, bigquery.QueryParameter{
		Name:  name,
		Value: value,
	})

	return fmt.Sprintf("@%s", name)
}

func (q *Query) buildSQL(dataset Dataset, b *sqlBuilder) string {
	if len(q.cols) == 0 {
		panic("should select cols for the query")
	}

	sql := "SELECT " + strings.Join(q.cols, ", ") + " FROM "
	if dataset != "" {
		sql += string(dataset) + "." + string(q.table)
	} else {
		sql += string(q.table)
	}
	if len(q.rootCond.filters) > 0 {
		sql += " WHERE " + q.rootCond.buildSQL(b)
	}
	if len(q.orders) > 0 {
		sql += " ORDER BY " + strings.Join(q.orders, ", ")
	}
	if q.limit > 0 {
		sql += fmt.Sprintf(" LIMIT %v", q.limit)
	}

	return sql
}

// String returns a human representation of the query to pretty print it.
func (q *Query) String() string {
	b := new(sqlBuilder)
	sql := q.buildSQL("", b)

	format := make([]string, len(b.params))
	for i, param := range b.params {
		format[i] = fmt.Sprintf(":%s => %#v", param.Name, param.Value)
	}
	if len(format) > 0 {
		sql += "; {" + strings.Join(format, ", ") + "}"
	}

	return sql
}

// Filter adds a new WHERE condition directly to the query. All the query conditions
// will be combined with the AND operator.
func (q *Query) Filter(filter string, value interface{}) *Query {
	q = q.Clone()
	q.rootCond = q.rootCond.Filter(filter, value)
	return q
}

// FilterCond adds a new child condition directly to the query. All the children
// query conditions will be combined with the AND operator.
func (q *Query) FilterCond(cond *Condition) *Query {
	q = q.Clone()
	q.rootCond = q.rootCond.FilterCond(cond)
	return q
}

// Checksum returns a checksum of the filters, conditions, table name, ... and other
// internal data of the collection that identifies it. It won't include the
// columns, so you can add or remove them without busting the checksums.
func (q *Query) Checksum(dataset Dataset, pageSize int32) uint32 {
	var checksum []string
	b := new(sqlBuilder)
	checksum = append(checksum, q.buildSQL(dataset, b))
	for _, param := range b.params {
		checksum = append(checksum, param.Name)
		checksum = append(checksum, fmt.Sprintf("%v", param.Value))
	}
	checksum = append(checksum, fmt.Sprintf("%d", pageSize))

	return crc32.ChecksumIEEE([]byte(strings.Join(checksum, "")))
}
