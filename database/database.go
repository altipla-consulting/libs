package database

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"

	log "github.com/sirupsen/logrus"
	"libs.altipla.consulting/errors"

	// Imports and registers the MySQL driver.
	_ "github.com/go-sql-driver/mysql"
)

// Database represents a reusable connection to a remote MySQL database.
type Database struct {
	sess  *sql.DB
	debug bool
}

// Open starts a new connection to a remote MySQL database using the provided credentials
func Open(credentials Credentials, options ...Option) (*Database, error) {
	db := new(Database)
	for _, option := range options {
		option(db)
	}

	if db.debug {
		log.WithField("credentials", credentials.String()).Debug("Open database connection")
	}

	var err error
	db.sess, err = sql.Open("mysql", credentials.String())
	if err != nil {
		return nil, errors.Wrapf(err, "cannot connect to mysql")
	}

	db.sess.SetMaxOpenConns(3)
	db.sess.SetMaxIdleConns(0)

	if err := db.sess.Ping(); err != nil {
		return nil, errors.Wrapf(err, "cannot ping mysql")
	}

	return db, nil
}

// Collection prepares a new collection using the table name of the model. It won't
// make any query, it only prepares the structs.
func (db *Database) Collection(model Model, opts ...CollectionOption) *Collection {
	c := newCollection(db, model)
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// Close the connection. You should not use a database after closing it, nor any
// of its generated collections.
func (db *Database) Close() {
	db.sess.Close()
}

// Exec runs a raw SQL query in the database and returns nothing. It is
// recommended to use Collections instead.
func (db *Database) Exec(query string, params ...interface{}) error {
	if db.debug {
		log.WithFields(log.Fields{
			"query":  query,
			"params": fmt.Sprintf("%#v", params),
		}).Debug("Exec SQL query")
	}

	_, err := db.sess.Exec(query, params...)
	return errors.Trace(err)
}

// QueryRow runs a raw SQL query in the database and returns the raw row from
// MySQL. It is recommended to use Collections instead.
func (db *Database) QueryRow(query string, params ...interface{}) *sql.Row {
	return db.sess.QueryRow(query, params...)
}

// Select fetchs a single row and loads the provided structure.
func (db *Database) Select(dest interface{}, query string, params ...interface{}) error {
	propsList, err := extractGenericProps(dest)
	if err != nil {
		return errors.Trace(err)
	}
	props := make(map[string]*Property)
	for _, prop := range propsList {
		props[prop.UnescapedName] = prop
	}

	rows, err := db.sess.Query(query, params...)
	if err != nil {
		return errors.Trace(err)
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return errors.Trace(err)
	}
	var pointers []interface{}
	for _, col := range cols {
		prop, ok := props[col]
		if !ok {
			return errors.Errorf("column in result not found in dest struct: %v", col)
		}

		pointers = append(pointers, prop.Pointer)
	}

	if !rows.Next() {
		return errors.Trace(ErrNoSuchEntity)
	}
	if err := rows.Scan(pointers...); err != nil {
		return errors.Trace(err)
	}
	if rows.Next() {
		return errors.Errorf("only one row expected in Select call, use LIMIT in your query to avoid more than one result")
	}
	if rows.Err() != nil {
		return errors.Trace(rows.Err())
	}

	return nil
}

// SelectAll fetchs the full list of rows and loads a pointer to a slice with them.
func (db *Database) SelectAll(dest interface{}, query string, params ...interface{}) error {
	v := reflect.ValueOf(dest)
	t := reflect.TypeOf(dest)

	if v.Kind() != reflect.Ptr {
		return errors.Errorf("pass a pointer to a slice to SelectAll")
	}
	v = v.Elem()
	t = t.Elem()
	if v.Kind() != reflect.Slice {
		return errors.Errorf("pass a slice to SelectAll")
	}

	example := reflect.New(t.Elem().Elem())
	globalProps, err := extractGenericProps(example.Interface())
	if err != nil {
		return errors.Trace(err)
	}

	rows, err := db.sess.Query(query, params...)
	if err != nil {
		return errors.Trace(err)
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return errors.Trace(err)
	}

	result := reflect.MakeSlice(t, 0, 0)
	for rows.Next() {
		model := reflect.New(t.Elem().Elem())

		props := make(map[string]*Property)
		for _, prop := range updateGenericProps(globalProps, model.Interface()) {
			props[prop.UnescapedName] = prop
		}
		var pointers []interface{}
		for _, col := range cols {
			prop, ok := props[col]
			if !ok {
				return errors.Errorf("column in result not found in dest struct: %v", col)
			}

			pointers = append(pointers, prop.Pointer)
		}

		if err := rows.Scan(pointers...); err != nil {
			return errors.Trace(err)
		}

		result = reflect.Append(result, model)
	}

	v.Set(result)

	return nil
}

type key int

const (
	keyTx = key(1)
)

type executor interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}

func (db *Database) executor(ctx context.Context) executor {
	tx, ok := ctx.Value(keyTx).(*sql.Tx)
	if ok {
		return tx
	}
	return db.sess
}

type TransactionalFn func(ctx context.Context) error

func (db *Database) Transaction(ctx context.Context, fn TransactionalFn) error {
	tx, err := db.sess.BeginTx(ctx, nil)
	if err != nil {
		return errors.Trace(err)
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	ctx = context.WithValue(ctx, keyTx, tx)

	if err := fn(ctx); err != nil {
		if err := tx.Rollback(); err != nil {
			return errors.Wrapf(err, "unable to rollback transaction")
		}

		return errors.Trace(err)
	}

	if err := tx.Commit(); err != nil {
		return errors.Trace(err)
	}

	return nil
}

// Option can be passed when opening a new connection to a database.
type Option func(db *Database)

// WithDebug is a database option that enables debug logging in the library.
func WithDebug(debug bool) Option {
	return func(db *Database) {
		db.debug = debug
	}
}
