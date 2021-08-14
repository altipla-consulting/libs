package rdb

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"libs.altipla.consulting/rdb/api"
)

func initTestbed(t *testing.T) *Database {
	db, err := Open("http://localhost:13000", "pkg-rdb", WithStrongConsistency(), WithLocalCreate())
	require.NoError(t, err)
	return db
}

type WrongModel struct {
	ModelTracking

	Foo time.Time
}

func (model *WrongModel) Collection() string {
	return "WrongModels"
}

func TestCheckTimeFields(t *testing.T) {
	db := initTestbed(t)

	require.PanicsWithValue(t, "do not use time.Time in models. Use rdb.Date or rdb.DateTime instead", func() {
		db.Collection(new(WrongModel))
	})
}

type WrongRecursiveModel struct {
	ModelTracking

	Foo FooChild1
}

func (model *WrongRecursiveModel) Collection() string {
	return "WrongRecursiveModels"
}

type FooChild1 struct {
	Foo FooChild2
}

type FooChild2 struct {
	Foo time.Time
}

func TestCheckTimeFieldsRecursive(t *testing.T) {
	db := initTestbed(t)

	require.PanicsWithValue(t, "do not use time.Time in models. Use rdb.Date or rdb.DateTime instead", func() {
		db.Collection(new(WrongRecursiveModel))
	})
}

type PrivateTimeModel struct {
	ModelTracking

	private time.Time
}

func (model *PrivateTimeModel) Collection() string {
	return "PrivateTimeModels"
}

func TestCheckTimeFieldsPrivateNotPanic(t *testing.T) {
	db := initTestbed(t)

	require.NotPanics(t, func() {
		db.Collection(new(PrivateTimeModel))
	})
}

type CorrectTimeModel struct {
	ModelTracking

	CreateTime DateTime
}

func (model *CorrectTimeModel) Collection() string {
	return "CorrectTimeModels"
}

func TestCheckTimeFieldsCorrectNotPanic(t *testing.T) {
	db := initTestbed(t)

	require.NotPanics(t, func() {
		db.Collection(new(CorrectTimeModel))
	})
}

type IndexResultModel struct {
	ModelTracking

	DisplayName string
}

func (model *IndexResultModel) Collection() string {
	return "IndexResultModels"
}

func TestCreateIndex(t *testing.T) {
	ctx := context.Background()
	db := initTestbed(t)

	index := Index{
		Maps: []string{`from foo in docs.Foo select new { foo.DisplayName }`},
	}
	require.NoError(t, db.CreateIndex(context.Background(), "FooIndex", index))

	var results []*IndexResultModel
	require.NoError(t, db.QueryIndex("FooIndex", new(IndexResultModel)).GetAll(ctx, &results))
}

type PatchModel struct {
	ModelTracking

	ID    string
	Value string
	Array []string
}

func (model *PatchModel) Collection() string {
	return "PatchModels"
}

func TestDatabasePatch(t *testing.T) {
	ctx := context.Background()
	db := initTestbed(t)
	collection := db.Collection(new(PatchModel))

	require.NoError(t, collection.DeleteEverything(ctx))

	foo1 := &PatchModel{
		ID:    "patch-models/1",
		Value: "foo1",
	}
	require.NoError(t, collection.Put(ctx, foo1))

	foo2 := &PatchModel{
		ID:    "patch-models/2",
		Value: "foo2",
	}
	require.NoError(t, collection.Put(ctx, foo2))

	foo3 := &PatchModel{
		ID:    "patch-models/3",
		Value: "foo3",
	}
	require.NoError(t, collection.Put(ctx, foo3))

	q := NewRQLQuery(`from PatchModels as m update { m.Value += $p1; }`)
	q.Set("p1", "b")
	op, err := db.Patch(ctx, q)
	require.NoError(t, err)
	require.NoError(t, op.WaitFor(ctx, 10*time.Millisecond))

	var models []*PatchModel
	require.NoError(t, collection.OrderByAlpha("id()").GetAll(ctx, &models))

	require.Len(t, models, 3)
	require.Equal(t, models[0].Value, "foo1b")
	require.Equal(t, models[1].Value, "foo2b")
	require.Equal(t, models[2].Value, "foo3b")
}

func TestDatabasePatchPushArray(t *testing.T) {
	ctx := context.Background()
	db := initTestbed(t)
	collection := db.Collection(new(PatchModel))

	require.NoError(t, collection.DeleteEverything(ctx))

	foo1 := &PatchModel{
		ID:    "patch-models/1",
		Array: []string{"foo1"},
	}
	require.NoError(t, collection.Put(ctx, foo1))

	foo2 := &PatchModel{
		ID:    "patch-models/2",
		Array: []string{"foo2"},
	}
	require.NoError(t, collection.Put(ctx, foo2))

	foo3 := &PatchModel{
		ID:    "patch-models/3",
		Array: []string{"foo3"},
	}
	require.NoError(t, collection.Put(ctx, foo3))

	q := NewRQLQuery(`from PatchModels as m where id() in ($ids) update { m.Array.push($value) }`)
	q.Set("ids", []string{"patch-models/1", "patch-models/3"})
	q.Set("value", "foo4")
	op, err := db.Patch(ctx, q)
	require.NoError(t, err)
	require.NoError(t, op.WaitFor(ctx, 10*time.Millisecond))

	var models []*PatchModel
	require.NoError(t, collection.OrderByAlpha("id()").GetAll(ctx, &models))

	require.Len(t, models, 3)
	require.Equal(t, models[0].Array, []string{"foo1", "foo4"})
	require.Equal(t, models[1].Array, []string{"foo2"})
	require.Equal(t, models[2].Array, []string{"foo3", "foo4"})
}

func TestDatabasePatchFilterArray(t *testing.T) {
	ctx := context.Background()
	db := initTestbed(t)
	collection := db.Collection(new(PatchModel))

	require.NoError(t, collection.DeleteEverything(ctx))

	foo1 := &PatchModel{
		ID:    "patch-models/1",
		Array: []string{"foo1", "foo4"},
	}
	require.NoError(t, collection.Put(ctx, foo1))

	foo2 := &PatchModel{
		ID:    "patch-models/2",
		Array: []string{"foo2"},
	}
	require.NoError(t, collection.Put(ctx, foo2))

	foo3 := &PatchModel{
		ID:    "patch-models/3",
		Array: []string{"foo3", "foo4"},
	}
	require.NoError(t, collection.Put(ctx, foo3))

	q := NewRQLQuery(`from PatchModels as m update { m.Array = m.Array.filter(item => item !== $value) }`)
	q.Set("value", "foo4")
	op, err := db.Patch(ctx, q)
	require.NoError(t, err)
	require.NoError(t, op.WaitFor(ctx, 10*time.Millisecond))

	var models []*PatchModel
	require.NoError(t, collection.OrderByAlpha("id()").GetAll(ctx, &models))

	require.Len(t, models, 3)
	require.Equal(t, models[0].Array, []string{"foo1"})
	require.Equal(t, models[1].Array, []string{"foo2"})
	require.Equal(t, models[2].Array, []string{"foo3"})
}

func TestDatabasePatchPushArrayNull(t *testing.T) {
	ctx := context.Background()
	db := initTestbed(t)
	collection := db.Collection(new(PatchModel))

	require.NoError(t, collection.DeleteEverything(ctx))

	foo := &PatchModel{
		ID: "patch-models/1",
	}
	require.NoError(t, collection.Put(ctx, foo))

	q := NewRQLQuery(`from PatchModels as m update { m.Array ? m.Array.push($value) : m.Array = [$value] }`)
	q.Set("value", "foo")
	op, err := db.Patch(ctx, q)
	require.NoError(t, err)
	require.NoError(t, op.WaitFor(ctx, 10*time.Millisecond))

	var model *PatchModel
	require.NoError(t, collection.First(ctx, &model))

	require.Equal(t, model.Array, []string{"foo"})
}

func TestDatabaseConfigureDefaultRevisions(t *testing.T) {
	ctx := context.Background()
	db := initTestbed(t)

	revs := &api.RevisionConfig{
		MinimumRevisionsToKeep:   3,
		MinimumRevisionAgeToKeep: api.Duration(3 * time.Hour),
		PurgeOnDelete:            true,
	}
	require.NoError(t, db.ConfigureDefaultRevisions(ctx, revs))

	desc, err := db.Descriptor(ctx)
	require.NoError(t, err)
	require.NotNil(t, desc.Revisions)
	require.NotNil(t, desc.Revisions.Default)
	require.EqualValues(t, desc.Revisions.Default.MinimumRevisionsToKeep, 3)
	require.EqualValues(t, desc.Revisions.Default.MinimumRevisionAgeToKeep, 3*time.Hour)
	require.True(t, desc.Revisions.Default.PurgeOnDelete)
	require.False(t, desc.Revisions.Default.Disabled)

	require.NoError(t, db.ConfigureDefaultRevisions(ctx, nil))

	desc, err = db.Descriptor(ctx)
	require.NoError(t, err)
	require.Nil(t, desc.Revisions.Default)
}
