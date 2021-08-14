package rdb

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

type FooCounterModel struct {
	ModelTracking

	ID string
}

func (model *FooCounterModel) Collection() string {
	return "FooCounterModels"
}

func initCounterTestbed(t *testing.T) *Database {
	ctx := context.Background()
	db := initTestbed(t)

	require.NoError(t, db.Collection(new(FooCounterModel)).DeleteEverything(ctx))

	return db
}

func TestCounterIncrement(t *testing.T) {
	ctx := context.Background()
	db := initCounterTestbed(t)
	collection := db.Collection(new(FooCounterModel))

	foo := &FooCounterModel{
		ID: "foo-counters/1",
	}
	require.NoError(t, collection.Put(ctx, foo))

	ctx, sess := db.NewSession(ctx)

	require.NoError(t, sess.Counter("foo-counters/1", "foo").Increment(ctx, 2))

	var other *FooCounterModel
	require.NoError(t, collection.Get(ctx, "foo-counters/1", &other, IncludeAllCounters()))

	counter := sess.Counter("foo-counters/1", "foo")
	require.EqualValues(t, counter.Value(), 2)
}

func TestCounterDecrement(t *testing.T) {
	ctx := context.Background()
	db := initCounterTestbed(t)
	collection := db.Collection(new(FooCounterModel))

	foo := &FooCounterModel{
		ID: "foo-counters/1",
	}
	require.NoError(t, collection.Put(ctx, foo))

	ctx, sess := db.NewSession(ctx)

	require.NoError(t, sess.Counter("foo-counters/1", "foo").Decrement(ctx, 2))

	var other *FooCounterModel
	require.NoError(t, collection.Get(ctx, "foo-counters/1", &other, IncludeAllCounters()))

	counter := sess.Counter("foo-counters/1", "foo")
	require.EqualValues(t, counter.Value(), -2)
}

func TestCounterDelete(t *testing.T) {
	ctx := context.Background()
	db := initCounterTestbed(t)
	collection := db.Collection(new(FooCounterModel))

	foo := &FooCounterModel{
		ID: "foo-counters/1",
	}
	require.NoError(t, collection.Put(ctx, foo))

	ctx, sess := db.NewSession(ctx)

	require.NoError(t, sess.Counter("foo-counters/1", "foo").Increment(ctx, 2))

	require.NoError(t, sess.Counter("foo-counters/1", "foo").Delete(ctx))

	var other *FooCounterModel
	require.NoError(t, collection.Get(ctx, "foo-counters/1", &other, IncludeAllCounters()))

	counter := sess.Counter("foo-counters/1", "foo")
	require.EqualValues(t, counter.Value(), 0)
}

func TestCounterEntityNotFound(t *testing.T) {
	ctx := context.Background()
	db := initCounterTestbed(t)

	ctx, sess := db.NewSession(ctx)

	require.EqualError(t, sess.Counter("foo-counters/1", "foo").Increment(ctx, 2), ErrNoSuchEntity.Error())
}
