package redis

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var db *Database

func initDB(t *testing.T) {
	db = Open("redis:6379", "test")
	db.FlushAllKeysFromDatabase()
}

func closeDB(t *testing.T) {
	require.NoError(t, db.Close())
}

type hashItem struct {
	StrField  string    `redis:"str_field"`
	IntField  int64     `redis:"int_field"`
	TimeField time.Time `redis:"time_field"`
	FreeField string
}

func (item *hashItem) IsRedisModel() bool {
	return true
}

func TestGetNoSuchEntity(t *testing.T) {
	initDB(t)
	defer closeDB(t)
	hash := db.Hash("foo-hash", new(hashItem))

	require.EqualError(t, hash.Get("foo", new(hashItem)), ErrNoSuchEntity.Error())
}

func TestGet(t *testing.T) {
	initDB(t)
	defer closeDB(t)
	hash := db.Hash("foo-hash", new(hashItem))

	timeField, err := time.Date(2006, time.January, 2, 15, 4, 5, 0, time.UTC).MarshalText()
	require.NoError(t, err)
	fields := map[string]interface{}{
		"str_field":  "foo str field",
		"int_field":  "32",
		"time_field": timeField,
		"FreeField":  "free str field",
	}
	require.NoError(t, db.sess.HMSet("test:foo-hash:foo", fields).Err())

	item := new(hashItem)
	require.NoError(t, hash.Get("foo", item))

	require.Equal(t, item.StrField, "foo str field")
	require.EqualValues(t, item.IntField, 32)
	require.Equal(t, item.TimeField, time.Date(2006, time.January, 2, 15, 4, 5, 0, time.UTC))
	require.Equal(t, item.FreeField, "free str field")
}

func TestPut(t *testing.T) {
	initDB(t)
	defer closeDB(t)
	hash := db.Hash("foo-hash", new(hashItem))

	item := &hashItem{
		StrField:  "foo str field",
		IntField:  32,
		TimeField: time.Date(2006, time.January, 2, 15, 4, 5, 0, time.UTC),
		FreeField: "free str field",
	}
	require.NoError(t, hash.Put("foo", item))

	fields, err := db.sess.HMGet("test:foo-hash:foo", "str_field", "int_field", "time_field", "FreeField").Result()
	require.NoError(t, err)

	require.Equal(t, fields[0].(string), "foo str field")
	require.Equal(t, fields[1].(string), "32")
	require.Equal(t, fields[2].(string), "2006-01-02T15:04:05Z")
	require.Equal(t, fields[3].(string), "free str field")
}

func TestPutGet(t *testing.T) {
	initDB(t)
	defer closeDB(t)
	hash := db.Hash("foo-hash", new(hashItem))

	save := &hashItem{
		StrField:  "foo str field",
		IntField:  32,
		TimeField: time.Date(2006, time.January, 2, 15, 4, 5, 0, time.UTC),
		FreeField: "free str field",
	}
	require.NoError(t, hash.Put("foo", save))

	item := new(hashItem)
	require.NoError(t, hash.Get("foo", item))

	require.Equal(t, item.StrField, "foo str field")
	require.EqualValues(t, item.IntField, 32)
	require.Equal(t, item.TimeField, time.Date(2006, time.January, 2, 15, 4, 5, 0, time.UTC))
	require.Equal(t, item.FreeField, "free str field")
}

func TestPutMask(t *testing.T) {
	initDB(t)
	defer closeDB(t)
	hash := db.Hash("foo-hash", new(hashItem))

	item := &hashItem{
		StrField:  "foo str field",
		IntField:  32,
		TimeField: time.Date(2006, time.January, 2, 15, 4, 5, 0, time.UTC),
		FreeField: "free str field",
	}
	require.NoError(t, hash.Put("foo", item, Mask("str_field", "int_field")))

	fields, err := db.sess.HMGet("test:foo-hash:foo", "str_field", "int_field", "time_field", "FreeField").Result()
	require.NoError(t, err)

	require.Equal(t, fields[0].(string), "foo str field")
	require.Equal(t, fields[1].(string), "32")
	require.Nil(t, fields[2])
	require.Nil(t, fields[3])
}

func TestGetMask(t *testing.T) {
	initDB(t)
	defer closeDB(t)
	hash := db.Hash("foo-hash", new(hashItem))

	fields := map[string]interface{}{
		"str_field": "foo str field",
		"int_field": "32",
		"FreeField": "free str field",
	}
	require.NoError(t, db.sess.HMSet("test:foo-hash:foo", fields).Err())

	item := new(hashItem)
	require.NoError(t, hash.Get("foo", item, Mask("str_field", "FreeField")))

	require.Equal(t, item.StrField, "foo str field")
	require.Zero(t, item.IntField)
	require.Zero(t, item.TimeField)
	require.Equal(t, item.FreeField, "free str field")
}

func TestDelete(t *testing.T) {
	initDB(t)
	defer closeDB(t)
	hash := db.Hash("foo-hash", new(hashItem))

	item := &hashItem{
		StrField:  "foo str field",
		IntField:  32,
		TimeField: time.Date(2006, time.January, 2, 15, 4, 5, 0, time.UTC),
		FreeField: "free str field",
	}
	require.NoError(t, hash.Put("foo", item))

	require.NoError(t, hash.Delete("foo"))

	require.EqualError(t, hash.Get("foo", item), ErrNoSuchEntity.Error())
}
