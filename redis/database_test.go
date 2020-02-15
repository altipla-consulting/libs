package redis

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

var db *Database

func initDB(t *testing.T) {
	db = Open("redis:6379", "test")
	require.NoError(t, db.FlushAllKeysFromDatabase())
}

func closeDB(t *testing.T) {
	require.NoError(t, db.Close())
}

func TestTransaction(t *testing.T) {
	initDB(t)
	defer closeDB(t)

	foo := db.StringKV("foo")

	err := db.Transaction(context.Background(), func(ctx context.Context) error {
		require.NoError(t, foo.Set(context.Background(), "bar"))

		v, err := foo.Get(context.Background())
		require.NoError(t, err)
		require.Equal(t, v, "bar")

		require.NoError(t, foo.Set(ctx, "baz"))

		v, err = foo.Get(context.Background())
		require.NoError(t, err)
		require.Equal(t, v, "bar")

		return nil
	})
	require.NoError(t, err)

	v, err := foo.Get(context.Background())
	require.NoError(t, err)
	require.Equal(t, v, "baz")
}
