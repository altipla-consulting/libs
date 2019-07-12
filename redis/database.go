package redis

import (
	"context"
	"fmt"

	"github.com/go-redis/redis"
)

// Database keeps a connection to a Redis server.
type Database struct {
	app        string
	directSess *redis.Client
}

// Open a new database connection to the remote Redis server.
func Open(hostname, applicationName string) *Database {
	return &Database{
		app:        applicationName,
		directSess: redis.NewClient(&redis.Options{Addr: hostname}),
	}
}

// Close the connection to the remote database.
func (db *Database) Close() error {
	return db.directSess.Close()
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

// FlushAllKeysFromDatabase is exposed as a simple way for tests to reset the local
// database. It is not intended to be run in production. It will clean up all the keys
// of the whole database and leav an empty canvas to fill again.
func (db *Database) FlushAllKeysFromDatabase() error {
	return db.directSess.FlushAll().Err()
}

// PubSub returns an entrypoint to a PubSub queue in redisIt can be used to publish
// and receive protobuf messages.
func (db *Database) PubSub(name string) *PubSub {
	return &PubSub{
		db:   db,
		name: fmt.Sprintf("%s:%s", db.app, name),
	}
}

// Hash returns a hash instance that stores full models as individual hash keys.
func (db *Database) Hash(name string, model Model) *Hash {
	props, err := extractModelProps(model)
	if err != nil {
		panic(err)
	}

	return &Hash{
		db:    db,
		name:  fmt.Sprintf("%s:%s", db.app, name),
		props: props,
	}
}

// Cmdable returns a reference to the direct Redis database connection. If context
// is inside a transaction it will return the Tx object instead. Both objects has
// the same Cmdable interface of the third party library.
func (db *Database) Cmdable(ctx context.Context) redis.Cmdable {
	return db.directSess
}
