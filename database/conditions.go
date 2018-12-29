package database

import (
	"fmt"
	"reflect"
	"strings"
)

// Condition should be implemented by any generic SQL condition we can apply to collections.
type Condition interface {
	// SQL returns the portion of the code that will be merged inside the WHERE
	// query. It can have placeholders with "?" to fill them apart.
	SQL() string

	// Values returns the list of placeholders values we should fill.
	Values() []interface{}
}

// Filter applies a new simple filter to the collection. There are multiple types
// of simple filters depending on the SQL you pass to it:
//
//   Filter("foo", "bar")
//   Filter("foo >", 3)
//   Filter("foo LIKE", "%bar%")
//   Filter("DATE_DIFF(?, mycolumn) > 30", time.Now())
func Filter(sql string, value interface{}) Condition {
	var queryValues []interface{}
	if !strings.Contains(sql, " ") {
		sql = fmt.Sprintf("%s = ?", sql)
		queryValues = []interface{}{value}
	} else if strings.Contains(sql, " IN") {
		v := reflect.ValueOf(value)

		placeholders := make([]string, v.Len())
		for i := 0; i < v.Len(); i++ {
			placeholders[i] = "?"
			queryValues = append(queryValues, v.Index(i).Interface())
		}
		sql = fmt.Sprintf("%s (%s)", sql, strings.Join(placeholders, ", "))

	} else if !strings.Contains(sql, "?") {
		sql = fmt.Sprintf("%s ?", sql)
		queryValues = []interface{}{value}
	} else {
		queryValues = []interface{}{value}
	}

	return &sqlCondition{sql, queryValues}
}

// CompareJSON creates a new condition that checks if a value inside a JSON
// object of a column is equal to the provided value.
func CompareJSON(column, path string, value interface{}) Condition {
	return &sqlCondition{
		sql:    fmt.Sprintf("JSON_EXTRACT(%s, '%s') = ?", column, path),
		values: []interface{}{value},
	}
}

// FilterExists checks if a subquery matches for each row before accepting it. It will use
// the join SQL statement as an additional filter to those ones both queries have to join the
// rows of two queries. Not having a join statement will throw a panic.
//
// No external parameters are allowed in the join statement because they can be supplied through
// normal filters in both collections. Limit yourself to relate both tables to make the FilterExists
// call useful.
//
// You can alias both collections to use shorter names in the statement. It is recommend to
// always use aliases when referring to the columns in the join statement.
func FilterExists(sub *Collection, join string) Condition {
	if join == "" {
		panic("join SQL statement is required to FilterExists")
	}

	sub = sub.Clone().FilterCond(&sqlCondition{join, nil})
	b := &sqlBuilder{
		table:      sub.model.TableName(),
		conditions: sub.conditions,
		props:      sub.props,
		limit:      sub.limit,
		offset:     sub.offset,
		orders:     sub.orders,
		alias:      sub.alias,
	}
	sql, values := b.SelectSQLCols("NULL")
	return &sqlCondition{fmt.Sprintf("EXISTS (%s)", sql), values}
}

type sqlCondition struct {
	sql    string
	values []interface{}
}

func (cond *sqlCondition) SQL() string {
	return cond.sql
}

func (cond *sqlCondition) Values() []interface{} {
	return cond.values
}

// And applies an AND operation between each of the children conditions.
func And(children []Condition) Condition {
	if len(children) == 0 {
		return new(sqlCondition)
	}

	sql := make([]string, len(children))
	values := []interface{}{}
	for i, child := range children {
		sql[i] = child.SQL()
		values = append(values, child.Values()...)
	}

	return &sqlCondition{
		sql:    "(" + strings.Join(sql, " AND ") + ")",
		values: values,
	}
}

// Or applies an OR operation between each of the children conditions.
func Or(children []Condition) Condition {
	if len(children) == 0 {
		return new(sqlCondition)
	}

	sql := make([]string, len(children))
	values := []interface{}{}
	for i, child := range children {
		sql[i] = child.SQL()
		values = append(values, child.Values()...)
	}

	return &sqlCondition{
		sql:    "(" + strings.Join(sql, " OR ") + ")",
		values: values,
	}
}

// EscapeLike escapes a value to insert it in a LIKE query without unexpected wildcards.
// After using this function to clean the value you can add the wildcards you need
// to the query.
func EscapeLike(str string) string {
	str = strings.Replace(str, "%", `\%`, -1)
	str = strings.Replace(str, "_", `\_`, -1)
	return str
}

// FilterIsNil filter rows with with NULL in the column.
func FilterIsNil(column string) Condition {
	return &sqlCondition{
		sql: fmt.Sprintf("%s IS NULL", column),
	}
}

// FilterIsNotNil filter rows with with something other than NULL in the column.
func FilterIsNotNil(column string) Condition {
	return &sqlCondition{
		sql: fmt.Sprintf("%s IS NOT NULL", column),
	}
}
