package database

import (
	"context"
	"database/sql"
	"fmt"
	"hash/crc32"
	"reflect"
	"strings"

	"github.com/altipla-consulting/errors"
	log "github.com/sirupsen/logrus"
)

type CollectionOption func(c *Collection)

// Collection represents a table. You can apply further filters and operations
// to the collection and then query it with one of our read methods (Get, GetAll, ...)
// or use it to store new items (Put).
type Collection struct {
	db            *Database
	conditions    []Condition
	orders        []string
	offset, limit int64
	model         Model
	props         []*Property
	alias         string
	h             *hooker
}

func newCollection(db *Database, model Model) *Collection {
	props, err := extractModelProps(model)
	if err != nil {
		panic(err)
	}

	c := &Collection{
		db:    db,
		model: model,
		props: props,
		h:     new(hooker),
	}

	return c
}

// Clone returns a new collection with the same filters and configuration of
// the original one.
func (c *Collection) Clone() *Collection {
	return &Collection{
		db:         c.db,
		conditions: c.conditions,
		orders:     c.orders,
		offset:     c.offset,
		limit:      c.limit,
		model:      c.model,
		props:      c.props,
		alias:      c.alias,
		h:          c.h,
	}
}

// Alias changes the name of the table in the SQL query. It is useful in combination
// with FilterExists() to have a stable name for the tables that should be filtered.
func (c *Collection) Alias(alias string) *Collection {
	c.alias = alias
	return c
}

// Get retrieves the model matching the collection filters and the model primary key.
// If no model is found ErrNoSuchEntity will be returned and the model won't be touched.
func (c *Collection) Get(ctx context.Context, instance Model) error {
	modelProps := updateModelProps(c.props, instance)
	b := &sqlBuilder{
		table:      c.model.TableName(),
		conditions: c.conditions,
		alias:      c.alias,
		props:      modelProps,
	}

	for _, prop := range modelProps {
		if prop.PrimaryKey {
			b.conditions = append(b.conditions, Filter(prop.UnescapedName, prop.Value))
		}
	}

	statement, values := b.SelectSQL()
	if c.db.debug {
		log.Println("database [Get]:", statement)
	}

	var pointers []interface{}
	for _, prop := range modelProps {
		pointers = append(pointers, prop.Pointer)
	}
	if err := c.db.executor(ctx).QueryRowContext(ctx, statement, values...).Scan(pointers...); err != nil {
		if err == sql.ErrNoRows {
			return ErrNoSuchEntity
		}

		return errors.Trace(err)
	}

	modelProps = updateModelProps(c.props, instance)

	return instance.Tracking().AfterGet(modelProps)
}

// Put stores a new item of the collection. Any filter or limit of the
// collection won't be applied.
func (c *Collection) Put(ctx context.Context, instance Model) error {
	modelt := reflect.TypeOf(c.model)
	instancet := reflect.TypeOf(instance)
	if modelt != instancet {
		return errors.Errorf("expected instance of %s and got a instance of %s", modelt, instancet)
	}

	if h, ok := instance.(OnBeforePutHooker); ok {
		if err := h.OnBeforePutHook(); err != nil {
			return errors.Trace(err)
		}
	}

	b := &sqlBuilder{
		table: c.model.TableName(),
	}
	modelProps := updateModelProps(c.props, instance)

	var q string
	var values []interface{}
	if instance.Tracking().IsInserted() {
		b.conditions = append(b.conditions, c.conditions...)
		b.conditions = append(b.conditions, Filter("revision", instance.Tracking().StoredRevision()))

		for _, prop := range modelProps {
			if prop.PrimaryKey {
				b.conditions = append(b.conditions, Filter(prop.UnescapedName, prop.Value))
				continue
			}

			if prop.OmitEmpty && isZero(prop.Value) {
				continue
			}

			b.props = append(b.props, prop)
		}

		q, values = b.UpdateSQL()
	} else {
		for _, prop := range modelProps {
			if prop.OmitEmpty && isZero(prop.Value) {
				continue
			}

			b.props = append(b.props, prop)
		}

		q, values = b.InsertSQL()
	}

	if c.db.debug {
		log.WithFields(log.Fields{
			"query":  q,
			"params": fmt.Sprintf("%#v", values),
		}).Debug("Put instance")
	}

	result, err := c.db.executor(ctx).ExecContext(ctx, q, values...)
	if err != nil {
		return errors.Trace(err)
	}

	var pks int
	for _, prop := range modelProps {
		if prop.PrimaryKey {
			pks++
		}
	}
	if pks == 1 && !instance.Tracking().IsInserted() {
		id, err := result.LastInsertId()
		if err != nil {
			return fmt.Errorf("cannot get last inserted id: %v", err)
		}

		if id != 0 {
			for _, prop := range modelProps {
				if prop.PrimaryKey {
					if _, ok := prop.Value.(int64); ok {
						reflect.ValueOf(prop.Pointer).Elem().Set(reflect.ValueOf(id))
					}
				}
			}
		}
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("cannot get rows affected: %v", err)
	}
	if rows == 0 {
		return ErrConcurrentTransaction
	}

	if err := instance.Tracking().AfterPut(modelProps); err != nil {
		return errors.Trace(err)
	}

	if h, ok := instance.(OnAfterPutHooker); ok {
		if err := h.OnAfterPutHook(); err != nil {
			return errors.Trace(err)
		}
	}
	if err := c.h.runAfterPut(ctx, instance); err != nil {
		return errors.Trace(err)
	}

	return nil
}

// Filter applies a new simple filter to the collection. See the global Filter
// function for documentation.
func (c *Collection) Filter(sqlStmt string, value interface{}) *Collection {
	return c.FilterCond(Filter(sqlStmt, value))
}

// FilterIsNil applies a new NULL filter to the collection. See the global FilterIsNil
// function for documentation.
func (c *Collection) FilterIsNil(column string) *Collection {
	return c.FilterCond(FilterIsNil(column))
}

// FilterIsNotNil applies a new NOT NULL filter to the collection. See the
// global FilterIsNotNil function for documentation.
func (c *Collection) FilterIsNotNil(column string) *Collection {
	return c.FilterCond(FilterIsNotNil(column))
}

// FilterLikeIgnoreCase filter rows with a LIKE match with wildcards in both sides
// and ignoring the case of the values.
func (c *Collection) FilterLikeIgnoreCase(column, value string) *Collection {
	return c.FilterCond(FilterLikeIgnoreCase(column, value))
}

// FilterCond applies a generic condition to the collection. We have some helpers
// in this library to build conditions; and other libraries (like github.com/altipla-consulting/geo)
// can implement their own conditions too.
func (c *Collection) FilterCond(condition Condition) *Collection {
	if condition.SQL() == "" {
		return c
	}

	c.conditions = append(c.conditions, condition)
	return c
}

// Offset moves the initial position of the query. In combination with Limit
// it allows you to paginate the results.
func (c *Collection) Offset(offset int64) *Collection {
	c.offset = offset
	return c
}

// Limit adds a maximum number of results to the query.
func (c *Collection) Limit(limit int64) *Collection {
	c.limit = limit
	return c
}

// Order the collection of items. You can pass "column" for ascendent order or "-column"
// for descendent order. If you want to order by multiple columns call Order multiple
// times for each column, the will be joined.
func (c *Collection) Order(column string) *Collection {
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
		column = fmt.Sprintf("`%s` DESC", column[1:])
	} else {
		column = fmt.Sprintf("`%s` ASC", column)
	}

	c.orders = append(c.orders, column)
	return c
}

// OrderSorter sorts the collection of items. We have some helpers
// in this library to build sorters; and other libraries (like github.com/altipla-consulting/geo)
// can implement their own sorters too.
func (c *Collection) OrderSorter(sorter Sorter) *Collection {
	c.orders = append(c.orders, sorter.SQL())
	return c
}

// Delete removes a model from a collection. It uses the filters and the model
// primary key to find the row to remove, so it can return an error even if the
// PK exists when the filters do not match. Limits won't be applied but the offset
// of the collection will.
func (c *Collection) Delete(ctx context.Context, instance Model) error {
	b := &sqlBuilder{
		table:      c.model.TableName(),
		conditions: c.conditions,
		limit:      1,
		offset:     c.offset,
		alias:      c.alias,
	}
	modelProps := updateModelProps(c.props, instance)

	for _, prop := range modelProps {
		if prop.PrimaryKey {
			b.conditions = append(b.conditions, Filter(prop.UnescapedName, prop.Value))
		}
	}

	statement, values := b.DeleteSQL()
	if c.db.debug {
		log.Println("database [Delete]:", statement)
	}

	if _, err := c.db.executor(ctx).ExecContext(ctx, statement, values...); err != nil {
		return errors.Trace(err)
	}

	return instance.Tracking().AfterDelete(modelProps)
}

// Iterator returns a new iterator that can be used to extract models one by one in a loop.
// You should close the Iterator after you are done with it.
func (c *Collection) Iterator(ctx context.Context) (*Iterator, error) {
	b := &sqlBuilder{
		table:      c.model.TableName(),
		conditions: c.conditions,
		props:      c.props,
		limit:      c.limit,
		offset:     c.offset,
		orders:     c.orders,
		alias:      c.alias,
	}

	sqlStmt, values := b.SelectSQL()
	if c.db.debug {
		log.Println("database [Iterator]:", sqlStmt)
	}

	rows, err := c.db.executor(ctx).QueryContext(ctx, sqlStmt, values...)
	if err != nil {
		return nil, err
	}

	return &Iterator{rows, c.props}, nil
}

// GetAll receives a pointer to an empty slice of models and retrieves all the
// models that match the filters of the collection. Take care to avoid fetching large
// collections of models or you will run out of memory.
func (c *Collection) GetAll(ctx context.Context, models interface{}) error {
	v := reflect.ValueOf(models)
	t := reflect.TypeOf(models)

	if v.Kind() != reflect.Ptr {
		return errors.Errorf("pass a pointer to a slice to GetAll")
	}
	v = v.Elem()
	t = t.Elem()
	if v.Kind() != reflect.Slice {
		return errors.Errorf("pass a slice to GetAll")
	}

	modelt := reflect.TypeOf(c.model)
	if t.Elem() != modelt {
		return errors.Errorf("expected a slice of %s and got a slice of %s", modelt, t.Elem())
	}

	dest := reflect.MakeSlice(t, 0, 0)

	it, err := c.Iterator(ctx)
	if err != nil {
		return errors.Trace(err)
	}
	defer it.Close()

	for {
		model := reflect.New(t.Elem().Elem())
		if err := it.Next(model.Interface().(Model)); err != nil {
			if err == ErrDone {
				break
			}

			return errors.Trace(err)
		}

		dest = reflect.Append(dest, model)
	}

	v.Set(dest)

	return nil
}

// First returns the first model that matches the collection. If no one is found
// it will return ErrNoSuchEntity and it won't touch model.
func (c *Collection) First(ctx context.Context, instance Model) error {
	c = c.Limit(1)

	modelProps := updateModelProps(c.props, instance)
	b := &sqlBuilder{
		table:      c.model.TableName(),
		conditions: c.conditions,
		props:      modelProps,
		limit:      c.limit,
		offset:     c.offset,
		orders:     c.orders,
		alias:      c.alias,
	}

	statement, values := b.SelectSQL()
	if c.db.debug {
		log.Println("database [First]:", statement)
	}

	var pointers []interface{}
	for _, prop := range modelProps {
		pointers = append(pointers, prop.Pointer)
	}
	if err := c.db.executor(ctx).QueryRowContext(ctx, statement, values...).Scan(pointers...); err != nil {
		if err == sql.ErrNoRows {
			return ErrNoSuchEntity
		}

		return errors.Trace(err)
	}

	modelProps = updateModelProps(c.props, instance)

	return instance.Tracking().AfterGet(modelProps)
}

// Count queries the number of rows that the collection matches.
func (c *Collection) Count(ctx context.Context) (int64, error) {
	b := &sqlBuilder{
		table:      c.model.TableName(),
		conditions: c.conditions,
		alias:      c.alias,
	}

	sqlStmt, values := b.SelectSQLCols("COUNT(*)")
	if c.db.debug {
		log.Println("database [Count]:", sqlStmt)
	}

	var n int64
	if err := c.db.executor(ctx).QueryRowContext(ctx, sqlStmt, values...).Scan(&n); err != nil {
		return 0, err
	}

	return n, nil
}

// GetMulti queries multiple rows and return all of them in a list. Keys should
// be a list of primary keys to retrieve and models should be a pointer to an empty
// slice of models. If any of the primary keys is not found a MultiError will be returned.
// You can check the error type for MultiError and then loop over the list of errors, they will
// be in the same order as the keys and they will have nil's when the row is found. The result
// list will also have the same length as keys with nil's filled when the row is not found.
func (c *Collection) GetMulti(ctx context.Context, keys interface{}, models interface{}) error {
	v := reflect.ValueOf(models)
	t := reflect.TypeOf(models)
	keyst := reflect.TypeOf(keys)
	keysv := reflect.ValueOf(keys)

	if v.Kind() != reflect.Ptr {
		return errors.Errorf("pass a pointer to a slice of models to GetMulti")
	}
	v = v.Elem()
	t = t.Elem()
	if v.Kind() != reflect.Slice {
		return errors.Errorf("pass a slice of models to GetMulti")
	}

	if keyst.Kind() != reflect.Slice {
		return errors.Errorf("pass a slice of keys to GetMulti")
	}
	keyst = keyst.Elem()
	if keyst.Kind() != reflect.Int64 && keyst.Kind() != reflect.String {
		return errors.Errorf("pass a slice of string/int64 keys to GetMulti")
	}
	if keysv.Len() == 0 {
		return nil
	}

	var pk *Property
	for _, prop := range c.props {
		if prop.PrimaryKey {
			if pk != nil {
				return errors.Errorf("cannot use GetMulti with multiple primary keys")
			}

			pk = prop
		}
	}

	c = c.Filter(fmt.Sprintf("%s IN", pk.Name), keys)

	fetch := reflect.New(t)
	fetch.Elem().Set(reflect.MakeSlice(t, 0, 0))
	if err := c.GetAll(ctx, fetch.Interface()); err != nil {
		return errors.Trace(err)
	}

	stringKeys := map[string]reflect.Value{}
	intKeys := map[int64]reflect.Value{}

	var merr MultiError
	for i := 0; i < fetch.Elem().Len(); i++ {
		model := fetch.Elem().Index(i)

		pk := model.Elem().FieldByName(pk.Field).Interface()
		switch v := pk.(type) {
		case string:
			stringKeys[v] = model
		case int64:
			intKeys[v] = model

		default:
			panic("should not reach here")
		}
	}

	results := reflect.MakeSlice(t, 0, 0)
	for i := 0; i < keysv.Len(); i++ {
		switch v := keysv.Index(i).Interface().(type) {
		case string:
			model, ok := stringKeys[v]
			if !ok {
				merr = append(merr, ErrNoSuchEntity)
				results = reflect.Append(results, reflect.Zero(t.Elem()))
				continue
			}

			merr = append(merr, nil)
			results = reflect.Append(results, model)

		case int64:
			model, ok := intKeys[v]
			if !ok {
				merr = append(merr, ErrNoSuchEntity)
				results = reflect.Append(results, reflect.Zero(t.Elem()))
				continue
			}

			merr = append(merr, nil)
			results = reflect.Append(results, model)

		default:
			panic("should not reach here")
		}
	}

	v.Set(results)

	if merr.HasError() {
		return merr
	}
	return nil
}

// Truncate removes every single row of a table. It also resets any autoincrement
// value it may have to the value "1".
func (c *Collection) Truncate(ctx context.Context) error {
	if _, err := c.db.executor(ctx).ExecContext(ctx, fmt.Sprintf(`DELETE FROM %s`, c.model.TableName())); err != nil {
		return errors.Trace(err)
	}
	if _, err := c.db.executor(ctx).ExecContext(ctx, fmt.Sprintf(`ALTER TABLE %s AUTO_INCREMENT = 1`, c.model.TableName())); err != nil {
		return errors.Trace(err)
	}

	return nil
}

// FilterExists applies the global FilterExists condition to this collection. See
// the global function for documentation.
func (c *Collection) FilterExists(sub *Collection, join string) *Collection {
	return c.FilterCond(FilterExists(sub, join))
}

// FilterNotExists applies the global FilterNotExists condition to this collection. See
// the global function for documentation.
func (c *Collection) FilterNotExists(sub *Collection, join string) *Collection {
	return c.FilterCond(FilterNotExists(sub, join))
}

// Checksum returns a checksum of the filters, conditions, table name, ... and other
// internal data of the collection that identifies it. It won't include the
// columns, so you can add or remove them without busting the checksums.
func (c *Collection) Checksum() uint32 {
	var checksum []string
	checksum = append(checksum, c.model.TableName())
	checksum = append(checksum, c.orders...)
	for _, cond := range c.conditions {
		checksum = append(checksum, cond.SQL())
		for _, v := range cond.Values() {
			checksum = append(checksum, fmt.Sprintf("%v", v))
		}
	}
	checksum = append(checksum, c.alias)
	checksum = append(checksum, fmt.Sprintf("%d", c.limit))

	return crc32.ChecksumIEEE([]byte(strings.Join(checksum, "")))
}
