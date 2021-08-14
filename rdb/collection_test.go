package rdb

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"libs.altipla.consulting/rdb/api"
)

type FooCollectionModel struct {
	ModelTracking

	ID          string
	DisplayName string
}

func (model *FooCollectionModel) Collection() string {
	return "FooCollectionModels"
}

func initCollectionTestbed(t *testing.T) *Database {
	ctx := context.Background()
	db := initTestbed(t)

	require.NoError(t, db.Collection(new(FooCollectionModel)).DeleteEverything(ctx))
	require.NoError(t, db.Collection(new(NullTimeModel)).DeleteEverything(ctx))

	return db
}

func TestPutGetCycle(t *testing.T) {
	ctx := context.Background()
	db := initCollectionTestbed(t)
	collection := db.Collection(new(FooCollectionModel))

	foo := &FooCollectionModel{
		ID:          "foo-collections/3",
		DisplayName: "foo",
	}
	require.NoError(t, collection.Put(ctx, foo))

	var other *FooCollectionModel
	require.NoError(t, collection.Get(ctx, "foo-collections/3", &other))

	require.Equal(t, other.ID, "foo-collections/3")
	require.Equal(t, other.DisplayName, "foo")
}

func TestGetInclude(t *testing.T) {
	ctx := context.Background()
	db := initCollectionTestbed(t)
	collection := db.Collection(new(FooCollectionModel))

	foo3 := &FooCollectionModel{
		ID:          "foo-collections/3",
		DisplayName: "foo-collections/4",
	}
	require.NoError(t, collection.Put(ctx, foo3))

	foo4 := &FooCollectionModel{
		ID:          "foo-collections/4",
		DisplayName: "Foo-Dest",
	}
	require.NoError(t, collection.Put(ctx, foo4))

	ctx, sess := db.NewSession(ctx)

	var other3 *FooCollectionModel
	require.NoError(t, collection.Get(ctx, "foo-collections/3", &other3, Include("DisplayName")))

	require.Equal(t, other3.ID, "foo-collections/3")
	require.Equal(t, other3.DisplayName, "foo-collections/4")

	var other4 *FooCollectionModel
	require.NoError(t, sess.Load("foo-collections/4", &other4))

	require.Equal(t, other4.ID, "foo-collections/4")
	require.Equal(t, other4.DisplayName, "Foo-Dest")
}

func TestGetIncludeCounters(t *testing.T) {
	ctx := context.Background()
	db := initCollectionTestbed(t)
	collection := db.Collection(new(FooCollectionModel))

	foo := &FooCollectionModel{
		ID: "foo-collections/1",
	}
	require.NoError(t, collection.Put(ctx, foo))

	ctx, sess := db.NewSession(ctx)

	require.NoError(t, sess.Counter("foo-collections/1", "foo").Increment(ctx, 2))

	var other *FooCollectionModel
	require.NoError(t, collection.Get(ctx, "foo-collections/1", &other, IncludeAllCounters()))

	counter := sess.Counter("foo-collections/1", "foo")
	require.EqualValues(t, counter.Value(), 2)
}

func TestGetEnforced(t *testing.T) {
	ctx := context.Background()
	db := initCollectionTestbed(t)
	collection := db.Collection(new(FooCollectionModel))

	foo := &FooCollectionModel{
		ID:          "foo-collections/3",
		DisplayName: "foo",
	}
	require.NoError(t, collection.Put(ctx, foo))

	bar := &FooCollectionModel{
		ID:          "foo-collections/4",
		DisplayName: "bar",
	}
	require.NoError(t, collection.Put(ctx, bar))

	enforced := collection.Enforce(Enforcer{
		Model: func(model Model) bool {
			return model.(*FooCollectionModel).DisplayName != "foo"
		},
		Query: func(q *Query) *Query {
			return q.Filter("DisplayName !=", "foo")
		},
	})

	other := new(FooCollectionModel)
	require.EqualError(t, enforced.Get(ctx, "foo-collections/3", &other), ErrNoSuchEntity.Error())
	require.Nil(t, other)

	require.NoError(t, enforced.Get(ctx, "foo-collections/4", &other))
	require.NotNil(t, other)
}

func TestGetNotNilPointer(t *testing.T) {
	ctx := context.Background()
	db := initCollectionTestbed(t)
	collection := db.Collection(new(FooCollectionModel))

	foo := &FooCollectionModel{
		ID: "foo-collections/3",
	}
	require.NoError(t, collection.Put(ctx, foo))

	other := new(FooCollectionModel)
	require.NoError(t, collection.Get(ctx, "foo-collections/3", &other))
}

func TestGetNotNilPointerNoSuchEntity(t *testing.T) {
	ctx := context.Background()
	db := initCollectionTestbed(t)
	collection := db.Collection(new(FooCollectionModel))

	foo := new(FooCollectionModel)
	require.EqualError(t, collection.Get(ctx, "foo-collections/30", &foo), ErrNoSuchEntity.Error())
}

func TestGetNotFoundLeavesNilPointer(t *testing.T) {
	ctx := context.Background()
	db := initCollectionTestbed(t)
	collection := db.Collection(new(FooCollectionModel))

	var foo *FooCollectionModel
	require.EqualError(t, collection.Get(ctx, "foo-collections/30", &foo), ErrNoSuchEntity.Error())
	require.Nil(t, foo)
}

func TestGetNotFoundLeavesEntityUnchanged(t *testing.T) {
	ctx := context.Background()
	db := initCollectionTestbed(t)
	collection := db.Collection(new(FooCollectionModel))

	foo := &FooCollectionModel{
		DisplayName: "foo",
	}
	require.EqualError(t, collection.Get(ctx, "foo-collections/30", &foo), ErrNoSuchEntity.Error())
	require.NotNil(t, foo)
	require.Equal(t, foo.DisplayName, "foo")
}

func TestGetMulti(t *testing.T) {
	ctx := context.Background()
	db := initCollectionTestbed(t)
	collection := db.Collection(new(FooCollectionModel))

	foo1 := &FooCollectionModel{
		ID:          "foo-collections/1",
		DisplayName: "foo1",
	}
	require.NoError(t, collection.Put(ctx, foo1))
	foo2 := &FooCollectionModel{
		ID:          "foo-collections/2",
		DisplayName: "foo2",
	}
	require.NoError(t, collection.Put(ctx, foo2))

	var results []*FooCollectionModel
	require.NoError(t, collection.GetMulti(ctx, []string{"foo-collections/1", "foo-collections/2"}, &results))

	require.Len(t, results, 2)
	require.Equal(t, results[0].ID, "foo-collections/1")
	require.Equal(t, results[0].DisplayName, "foo1")
	require.Equal(t, results[1].ID, "foo-collections/2")
	require.Equal(t, results[1].DisplayName, "foo2")
}

func TestGetMultiDifferentOrder(t *testing.T) {
	ctx := context.Background()
	db := initCollectionTestbed(t)
	collection := db.Collection(new(FooCollectionModel))

	foo1 := &FooCollectionModel{
		ID:          "foo-collections/1",
		DisplayName: "foo1",
	}
	require.NoError(t, collection.Put(ctx, foo1))
	foo2 := &FooCollectionModel{
		ID:          "foo-collections/2",
		DisplayName: "foo2",
	}
	require.NoError(t, collection.Put(ctx, foo2))

	var results []*FooCollectionModel
	require.NoError(t, collection.GetMulti(ctx, []string{"foo-collections/2", "foo-collections/1"}, &results))

	require.Len(t, results, 2)
	require.Equal(t, results[0].ID, "foo-collections/2")
	require.Equal(t, results[0].DisplayName, "foo2")
	require.Equal(t, results[1].ID, "foo-collections/1")
	require.Equal(t, results[1].DisplayName, "foo1")
}

func TestGetMultiNoSuchEntityNone(t *testing.T) {
	ctx := context.Background()
	db := initCollectionTestbed(t)
	collection := db.Collection(new(FooCollectionModel))

	var results []*FooCollectionModel
	err := collection.GetMulti(ctx, []string{"foo-collections/1", "foo-collections/3"}, &results)
	require.EqualError(t, err, "rdb: no such entity; rdb: no such entity")

	var merr MultiError
	require.True(t, errors.As(err, &merr))

	require.Len(t, merr, 2)
	require.EqualError(t, merr[0], ErrNoSuchEntity.Error())
	require.EqualError(t, merr[1], ErrNoSuchEntity.Error())
}

func TestGetMultiNoSuchEntity(t *testing.T) {
	ctx := context.Background()
	db := initCollectionTestbed(t)
	collection := db.Collection(new(FooCollectionModel))

	foo := &FooCollectionModel{
		ID: "foo-collections/1",
	}
	require.NoError(t, collection.Put(ctx, foo))

	var results []*FooCollectionModel
	err := collection.GetMulti(ctx, []string{"foo-collections/1", "foo-collections/3"}, &results)
	require.EqualError(t, err, "<nil>; rdb: no such entity")

	var merr MultiError
	require.True(t, errors.As(err, &merr))

	require.Len(t, merr, 2)
	require.Nil(t, merr[0])
	require.EqualError(t, merr[1], ErrNoSuchEntity.Error())
}

func TestDelete(t *testing.T) {
	ctx := context.Background()
	db := initCollectionTestbed(t)
	collection := db.Collection(new(FooCollectionModel))

	foo := &FooCollectionModel{
		ID: "foo-collections/3",
	}
	require.NoError(t, collection.Put(ctx, foo))

	require.NoError(t, collection.Delete(ctx, foo))

	var other *FooCollectionModel
	require.EqualError(t, collection.Get(ctx, "foo-collections/3", &other), ErrNoSuchEntity.Error())
}

func TestConcurrentUpdates(t *testing.T) {
	ctx := context.Background()
	db := initCollectionTestbed(t)
	collection := db.Collection(new(FooCollectionModel))

	foo := &FooCollectionModel{
		ID:          "foo-collections/3",
		DisplayName: "foo",
	}
	require.NoError(t, collection.Put(ctx, foo))

	var first *FooCollectionModel
	require.NoError(t, collection.Get(ctx, "foo-collections/3", &first))

	var second *FooCollectionModel
	require.NoError(t, collection.Get(ctx, "foo-collections/3", &second))

	first.DisplayName = "foo-first"
	require.NoError(t, collection.Put(ctx, first))

	second.DisplayName = "foo-second"
	require.EqualError(t, collection.Put(ctx, second), ErrConcurrentTransaction.Error())
}

func TestConcurrentCreates(t *testing.T) {
	ctx := context.Background()
	db := initCollectionTestbed(t)
	collection := db.Collection(new(FooCollectionModel))

	bar := &FooCollectionModel{
		ID:          "foo-collections/3",
		DisplayName: "foo",
	}
	require.NoError(t, collection.Put(ctx, bar))

	second := &FooCollectionModel{
		ID:          "foo-collections/3",
		DisplayName: "bar",
	}
	require.EqualError(t, collection.Put(ctx, second), ErrConcurrentTransaction.Error())

	var saved *FooCollectionModel
	require.NoError(t, collection.Get(ctx, "foo-collections/3", &saved))
	require.Equal(t, saved.DisplayName, "foo")
}

func TestTryGetSuccessful(t *testing.T) {
	ctx := context.Background()
	db := initCollectionTestbed(t)
	collection := db.Collection(new(FooCollectionModel))

	foo := &FooCollectionModel{
		ID:          "foo-collections/3",
		DisplayName: "Foo",
	}
	require.NoError(t, collection.Put(ctx, foo))

	var other *FooCollectionModel
	require.NoError(t, collection.TryGet(ctx, "foo-collections/3", &other))

	require.NotNil(t, other)
	require.Equal(t, other.DisplayName, "Foo")
}

func TestTryGetNoSuchEntity(t *testing.T) {
	ctx := context.Background()
	db := initCollectionTestbed(t)
	collection := db.Collection(new(FooCollectionModel))

	var other *FooCollectionModel
	require.NoError(t, collection.TryGet(ctx, "foo-collections/30", &other))

	require.Nil(t, other)
}

func TestMultiplePut(t *testing.T) {
	db := initCollectionTestbed(t)
	collection := db.Collection(new(FooCollectionModel))
	ctx, sess := db.NewSession(context.Background())

	foo := &FooCollectionModel{
		ID: "foo-collections/3",
	}
	require.NoError(t, collection.Put(ctx, foo))

	bar := &FooCollectionModel{
		ID: "foo-collections/4",
	}
	require.NoError(t, collection.Put(ctx, bar))

	require.Empty(t, foo.ChangeVector())
	require.Empty(t, bar.ChangeVector())

	require.NoError(t, sess.SaveChanges(ctx))

	require.NotEmpty(t, foo.ChangeVector())
	require.NotEmpty(t, bar.ChangeVector())
}

func TestCollectionConfigureRevisions(t *testing.T) {
	ctx := context.Background()
	db := initTestbed(t)
	collection := db.Collection(new(FooCollectionModel))

	revs := &api.RevisionConfig{
		MinimumRevisionsToKeep:   3,
		MinimumRevisionAgeToKeep: api.Duration(3 * time.Hour),
		PurgeOnDelete:            true,
	}
	require.NoError(t, collection.ConfigureRevisions(ctx, revs))

	desc, err := db.Descriptor(ctx)
	require.NoError(t, err)
	require.NotNil(t, desc.Revisions)
	require.NotEmpty(t, desc.Revisions.Collections)
	require.Len(t, desc.Revisions.Collections, 1)
	rev := desc.Revisions.Collections["FooCollectionModels"]
	require.EqualValues(t, rev.MinimumRevisionsToKeep, 3)
	require.EqualValues(t, rev.MinimumRevisionAgeToKeep, 3*time.Hour)
	require.True(t, rev.PurgeOnDelete)
	require.False(t, rev.Disabled)

	require.NoError(t, collection.ConfigureRevisions(ctx, nil))

	desc, err = db.Descriptor(ctx)
	require.NoError(t, err)
	require.Empty(t, desc.Revisions.Collections)
}
