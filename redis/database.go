package redis

import (
	"fmt"

	"github.com/go-redis/redis"
)

// Database keeps a connection to a Redis server.
type Database struct {
	app  string
	sess *redis.Client
}

// Open a new database connection to the remote Redis server.
func Open(hostname, applicationName string) *Database {
	return &Database{
		app:  applicationName,
		sess: redis.NewClient(&redis.Options{Addr: hostname}),
	}
}

// Close the connection to the remote database.
func (db *Database) Close() error {
	return db.sess.Close()
}

func (db *Database) StringsSet(key string) *StringsSet {
	return &StringsSet{
		db:  db,
		key: fmt.Sprintf("%s:%s", db.app, key),
	}
}

func (db *Database) Int64Set(key string) *Int64Set {
	return &Int64Set{
		db:  db,
		key: fmt.Sprintf("%s:%s", db.app, key),
	}
}

func (db *Database) StringKV(key string) *StringKV {
	return &StringKV{
		db:  db,
		key: fmt.Sprintf("%s:%s", db.app, key),
	}
}

func (db *Database) Int32KV(key string) *Int32KV {
	return &Int32KV{
		db:  db,
		key: fmt.Sprintf("%s:%s", db.app, key),
	}
}

func (db *Database) Int64KV(key string) *Int64KV {
	return &Int64KV{
		db:  db,
		key: fmt.Sprintf("%s:%s", db.app, key),
	}
}

func (db *Database) ProtoKV(key string) *ProtoKV {
	return &ProtoKV{
		db:  db,
		key: fmt.Sprintf("%s:%s", db.app, key),
	}
}

func (db *Database) ProtoHash(key string) *ProtoHash {
	return &ProtoHash{
		db:  db,
		key: fmt.Sprintf("%s:%s", db.app, key),
	}
}

func (db *Database) Counters(key string) *Counters {
	return &Counters{
		db:  db,
		key: fmt.Sprintf("%s:%s", db.app, key),
	}
}

func (db *Database) BooleanKV(key string) *BooleanKV {
	return &BooleanKV{
		db:  db,
		key: fmt.Sprintf("%s:%s", db.app, key),
	}
}

func (db *Database) TimeKV(key string) *TimeKV {
	return &TimeKV{
		db:  db,
		key: fmt.Sprintf("%s:%s", db.app, key),
	}
}

func (db *Database) ProtoList(key string) *ProtoList {
	return &ProtoList{
		db:  db,
		key: fmt.Sprintf("%s:%s", db.app, key),
	}
}

func (db *Database) StringsList(key string) *StringsList {
	return &StringsList{
		db:  db,
		key: fmt.Sprintf("%s:%s", db.app, key),
	}
}

// DirectClient returns the underlying client of go-redis to call advanced
// methods not exposed through this library. Please consider to add the functionality
// here after it's tested to improve all the users of the library.
func (db *Database) DirectClient() *redis.Client {
	return db.sess
}

// FlushAllKeysFromDatabase is exposed as a simple way for tests to reset the local
// database. It is not intended to be run in production. It will clean up all the keys
// of the whole database and leav an empty canvas to fill again.
func (db *Database) FlushAllKeysFromDatabase() error {
	return db.sess.FlushAll().Err()
}

// PubSub returns an entrypoint to a PubSub queue in redisIt can be used to publish
// and receive protobuf messages.
func (db *Database) PubSub(name string) *PubSub {
	return &PubSub{
		db:   db,
		name: fmt.Sprintf("%s:%s", db.app, name),
	}
}
