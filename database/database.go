package database

import (
	"database/sql"
	"fmt"
	"log"

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
		log.Println("database [Open]:", credentials)
	}

	var err error
	db.sess, err = sql.Open("mysql", credentials.String())
	if err != nil {
		return nil, fmt.Errorf("database: cannot connect to mysql: %s", err)
	}

	db.sess.SetMaxOpenConns(3)
	db.sess.SetMaxIdleConns(0)

	if err := db.sess.Ping(); err != nil {
		return nil, fmt.Errorf("database: cannot ping mysql: %s", err)
	}

	return db, nil
}

// Collection prepares a new collection using the table name of the model. It won't
// make any query, it only prepares the structs.
func (db *Database) Collection(model Model) *Collection {
	return newCollection(db, model)
}

// Close the connection. You should not use a database after closing it, nor any
// of its generated collections.
func (db *Database) Close() {
	db.sess.Close()
}

// Exec runs a raw SQL query in the database and returns nothing. It is
// recommended to use Collections instead.
func (db *Database) Exec(query string, params ...interface{}) error {
	_, err := db.sess.Exec(query, params...)
	return err
}

// QueryRow runs a raw SQL query in the database and returns the raw row from
// MySQL. It is recommended to use Collections instead.
func (db *Database) QueryRow(query string, params ...interface{}) *sql.Row {
	return db.sess.QueryRow(query, params...)
}

// Option can be passed when opening a new connection to a database.
type Option func(db *Database)

// WithDebug is a database option that enables debug logging in the library.
func WithDebug(debug bool) Option {
	return func(db *Database) {
		db.debug = debug
	}
}
