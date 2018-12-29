package database

import (
	"fmt"
	"strings"
)

type sqlBuilder struct {
	table         string
	props         []*Property
	orders        []string
	conditions    []Condition
	limit, offset int64
	alias         string
}

func (b *sqlBuilder) cols() []string {
	var cols []string
	for _, prop := range b.props {
		cols = append(cols, prop.Name)
	}

	return cols
}

func (b *sqlBuilder) SelectSQL() (string, []interface{}) {
	return b.SelectSQLCols(b.cols()...)
}

func (b *sqlBuilder) SelectSQLCols(cols ...string) (string, []interface{}) {
	var conds []string
	var values []interface{}
	for _, cond := range b.conditions {
		conds = append(conds, cond.SQL())
		values = append(values, cond.Values()...)
	}

	sql := fmt.Sprintf(`SELECT %s FROM %s`, strings.Join(cols, ", "), b.table)
	if b.alias != "" {
		sql = fmt.Sprintf("%s AS %s", sql, b.alias)
	}

	if len(conds) > 0 {
		sql = fmt.Sprintf("%s WHERE %s", sql, strings.Join(conds, " AND "))
	}
	if len(b.orders) > 0 {
		sql = fmt.Sprintf("%s ORDER BY %s", sql, strings.Join(b.orders, ", "))
	}
	if b.limit > 0 {
		sql = fmt.Sprintf("%s LIMIT %d,%d", sql, b.offset, b.limit)
	}

	return sql, values
}

func (b *sqlBuilder) UpdateSQL() (string, []interface{}) {
	var values []interface{}

	var updates []string
	for _, prop := range b.props {
		updates = append(updates, fmt.Sprintf("%s = ?", prop.Name))
		values = append(values, prop.Value)
	}

	var conds []string
	for _, cond := range b.conditions {
		conds = append(conds, cond.SQL())
		values = append(values, cond.Values()...)
	}

	sql := fmt.Sprintf(`UPDATE %s SET %s WHERE %s`, b.table, strings.Join(updates, ", "), strings.Join(conds, " AND "))

	return sql, values
}

func (b *sqlBuilder) InsertSQL() (string, []interface{}) {
	var values []interface{}

	var placeholders []string
	for _, prop := range b.props {
		placeholders = append(placeholders, "?")
		values = append(values, prop.Value)
	}

	sql := fmt.Sprintf(`INSERT INTO %s(%s) VALUES(%s)`, b.table, strings.Join(b.cols(), ", "), strings.Join(placeholders, ", "))

	return sql, values
}

func (b *sqlBuilder) DeleteSQL() (string, []interface{}) {
	var conds []string
	var values []interface{}
	for _, cond := range b.conditions {
		conds = append(conds, cond.SQL())
		values = append(values, cond.Values()...)
	}

	sql := fmt.Sprintf(`DELETE FROM %s WHERE %s`, b.table, strings.Join(conds, " AND "))

	return sql, values
}

func (b *sqlBuilder) TruncateSQL() string {
	return fmt.Sprintf(`DELETE FROM %s`, b.table)
}

func (b *sqlBuilder) ResetAutoIncrementSQL() string {
	return fmt.Sprintf(`ALTER TABLE %s AUTO_INCREMENT = 1`, b.table)
}
