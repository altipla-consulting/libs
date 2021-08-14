package rdb

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type FooQueryModel struct {
	ModelTracking

	ID          string
	DisplayName string
	Alternative string
}

func (model *FooQueryModel) Collection() string {
	return "FooQueryModels"
}

type FooQueryNumericModel struct {
	ModelTracking

	ID     string
	Number int64
}

func (model *FooQueryNumericModel) Collection() string {
	return "FooQueryNumericModels"
}

type FooQueryIncludeModel struct {
	ModelTracking

	ID          string
	DisplayName string

	FooQuery string
}

func (model *FooQueryIncludeModel) Collection() string {
	return "FooQueryIncludeModels"
}

func initQueryTestbed(t *testing.T) *Database {
	db := initTestbed(t)
	ctx, sess := db.NewSession(context.Background())

	queryCollection := db.Collection(new(FooQueryModel))
	require.NoError(t, queryCollection.DeleteEverything(ctx))

	numericCollection := db.Collection(new(FooQueryNumericModel))
	require.NoError(t, numericCollection.DeleteEverything(ctx))

	includeCollection := db.Collection(new(FooQueryIncludeModel))
	require.NoError(t, includeCollection.DeleteEverything(ctx))

	indexCollection := db.Collection(new(FooQueryIndexModel))
	require.NoError(t, indexCollection.DeleteEverything(ctx))

	foo := &FooQueryModel{
		ID:          "foo-queries/1",
		DisplayName: "Foo1",
		Alternative: "Alt1",
	}
	require.NoError(t, queryCollection.Put(ctx, foo))
	foo = &FooQueryModel{
		ID:          "foo-queries/2",
		DisplayName: "Foo2",
		Alternative: "Alt2",
	}
	require.NoError(t, queryCollection.Put(ctx, foo))
	foo = &FooQueryModel{
		ID:          "foo-queries/3",
		DisplayName: "Foo3",
		Alternative: "Alt3",
	}
	require.NoError(t, queryCollection.Put(ctx, foo))

	require.NoError(t, sess.SaveChanges(ctx))

	return db
}

func TestQuerySimpleFilter(t *testing.T) {
	ctx := context.Background()
	db := initQueryTestbed(t)
	collection := db.Collection(new(FooQueryModel))

	var result *FooQueryModel
	require.NoError(t, collection.Filter("DisplayName", "Foo2").First(ctx, &result))

	require.Equal(t, result.ID, "foo-queries/2")
	require.Equal(t, result.DisplayName, "Foo2")
}

func TestQueryFilterOperator(t *testing.T) {
	ctx := context.Background()
	db := initQueryTestbed(t)
	collection := db.Collection(new(FooQueryModel))

	var results []*FooQueryModel
	require.NoError(t, collection.Filter("DisplayName !=", "Foo2").GetAll(ctx, &results))

	require.Len(t, results, 2)
	require.Equal(t, results[0].ID, "foo-queries/1")
	require.Equal(t, results[1].ID, "foo-queries/3")
}

func TestQueryFilterLogicalOR(t *testing.T) {
	ctx := context.Background()
	db := initQueryTestbed(t)
	collection := db.Collection(new(FooQueryModel))

	var results []*FooQueryModel
	q := collection.FilterSub(
		Or(
			Filter("DisplayName", "Foo1"),
			Filter("DisplayName", "Foo3")))
	require.NoError(t, q.GetAll(ctx, &results))

	require.Len(t, results, 2)
	require.Equal(t, results[0].ID, "foo-queries/1")
	require.Equal(t, results[1].ID, "foo-queries/3")
}

func TestQueryFilterIn(t *testing.T) {
	ctx := context.Background()
	db := initQueryTestbed(t)
	collection := db.Collection(new(FooQueryModel))

	var results []*FooQueryModel
	require.NoError(t, collection.FilterIn("DisplayName", "Foo1", "Foo3").GetAll(ctx, &results))

	require.Len(t, results, 2)
	require.Equal(t, results[0].ID, "foo-queries/1")
	require.Equal(t, results[1].ID, "foo-queries/3")
}

func TestQueryCount(t *testing.T) {
	ctx := context.Background()
	db := initQueryTestbed(t)
	collection := db.Collection(new(FooQueryModel))

	n, err := collection.FilterIn("DisplayName", "Foo1", "Foo3").Count(ctx)
	require.NoError(t, err)

	require.EqualValues(t, n, 2)
}

func TestQueryCountFullCollection(t *testing.T) {
	ctx := context.Background()
	db := initQueryTestbed(t)
	collection := db.Collection(new(FooQueryModel))

	n, err := collection.Count(ctx)
	require.NoError(t, err)

	require.EqualValues(t, n, 3)
}

func TestQueryFilterEndsWith(t *testing.T) {
	ctx := context.Background()
	db := initQueryTestbed(t)
	collection := db.Collection(new(FooQueryModel))

	var results []*FooQueryModel
	require.NoError(t, collection.FilterEndsWith("DisplayName", "3").GetAll(ctx, &results))

	require.Len(t, results, 1)
	require.Equal(t, results[0].ID, "foo-queries/3")
}

func TestQuerySelect(t *testing.T) {
	ctx := context.Background()
	db := initQueryTestbed(t)
	collection := db.Collection(new(FooQueryModel))

	var results []*FooQueryModel
	require.NoError(t, collection.Select("Alternative").GetAll(ctx, &results))

	require.Len(t, results, 3)
	require.Equal(t, results[0].ID, "foo-queries/1")
	require.Empty(t, results[0].DisplayName)
	require.Equal(t, results[0].Alternative, "Alt1")
	require.Equal(t, results[1].ID, "foo-queries/2")
	require.Empty(t, results[1].DisplayName)
	require.Equal(t, results[1].Alternative, "Alt2")
	require.Equal(t, results[2].ID, "foo-queries/3")
	require.Empty(t, results[2].DisplayName)
	require.Equal(t, results[2].Alternative, "Alt3")
}

type projectModel struct {
	ID          string
	DisplayName string
	Alternative string
}

func TestQueryProjectGetAll(t *testing.T) {
	ctx := context.Background()
	db := initQueryTestbed(t)
	collection := db.Collection(new(FooQueryModel))

	var results []*projectModel
	require.NoError(t, collection.Project(new(projectModel), "Alternative").GetAll(ctx, &results))

	require.Len(t, results, 3)
	require.Equal(t, results[0].ID, "foo-queries/1")
	require.Empty(t, results[0].DisplayName)
	require.Equal(t, results[0].Alternative, "Alt1")
	require.Equal(t, results[1].ID, "foo-queries/2")
	require.Empty(t, results[1].DisplayName)
	require.Equal(t, results[1].Alternative, "Alt2")
	require.Equal(t, results[2].ID, "foo-queries/3")
	require.Empty(t, results[2].DisplayName)
	require.Equal(t, results[2].Alternative, "Alt3")
}

func TestQueryProjectFirst(t *testing.T) {
	ctx := context.Background()
	db := initQueryTestbed(t)
	collection := db.Collection(new(FooQueryModel))

	var result *projectModel
	require.NoError(t, collection.Project(new(projectModel), "Alternative").First(ctx, &result))

	require.NotNil(t, result)
	require.Equal(t, result.ID, "foo-queries/1")
	require.Empty(t, result.DisplayName)
	require.Equal(t, result.Alternative, "Alt1")
}

func TestQueryComplexConditionsRQL(t *testing.T) {
	db := initQueryTestbed(t)
	collection := db.Collection(new(FooQueryModel))

	q := collection.
		FilterSub(
			Or(
				Filter("ScheduleTime", nil),
				Filter("ScheduleTime <=", time.Date(2020, time.January, 2, 3, 4, 5, 6, time.UTC)),
			),
		).
		Filter("Published", true).
		Filter("Content !=", "")

	sql, params := q.RQL()
	require.Equal(t, sql, `from FooQueryModels where (ScheduleTime = $p0 or ScheduleTime <= $p1) and Published = $p2 and exact(Content != $p3)`)
	require.Equal(t, params, map[string]interface{}{
		"p0": nil,
		"p1": "2020-01-02T03:04:05.0000000Z",
		"p2": true,
		"p3": "",
	})
}

func TestQueryOrderDefaultAlpha(t *testing.T) {
	ctx := context.Background()
	db := initQueryTestbed(t)
	collection := db.Collection(new(FooQueryModel))

	foo10 := &FooQueryModel{
		ID:          "foo-queries/10",
		DisplayName: "Foo10",
		Alternative: "Alt10",
	}
	require.NoError(t, collection.Put(ctx, foo10))

	var models []*FooQueryModel
	require.NoError(t, collection.OrderBy("-DisplayName").GetAll(ctx, &models))

	require.Len(t, models, 4)
	require.Equal(t, models[0].ID, "foo-queries/3")
	require.Equal(t, models[1].ID, "foo-queries/2")
	require.Equal(t, models[2].ID, "foo-queries/10")
	require.Equal(t, models[3].ID, "foo-queries/1")
}

func TestQueryOrderAlphaNumeric(t *testing.T) {
	ctx := context.Background()
	db := initQueryTestbed(t)
	collection := db.Collection(new(FooQueryModel))

	foo10 := &FooQueryModel{
		ID:          "foo-queries/10",
		DisplayName: "Foo10",
		Alternative: "Alt10",
	}
	require.NoError(t, collection.Put(ctx, foo10))

	var models []*FooQueryModel
	require.NoError(t, collection.OrderByAlphaNumeric("-DisplayName").GetAll(ctx, &models))

	require.Len(t, models, 4)
	require.Equal(t, models[0].ID, "foo-queries/10")
	require.Equal(t, models[1].ID, "foo-queries/3")
	require.Equal(t, models[2].ID, "foo-queries/2")
	require.Equal(t, models[3].ID, "foo-queries/1")
}

func TestQueryOrderDefaultForNumericFields(t *testing.T) {
	db := initQueryTestbed(t)
	ctx, sess := db.NewSession(context.Background())
	collection := db.Collection(new(FooQueryNumericModel))

	foo1 := &FooQueryNumericModel{
		ID:     "foo-queries-numeric/1",
		Number: 1,
	}
	require.NoError(t, collection.Put(ctx, foo1))
	foo2 := &FooQueryNumericModel{
		ID:     "foo-queries-numeric/2",
		Number: 2,
	}
	require.NoError(t, collection.Put(ctx, foo2))
	foo3 := &FooQueryNumericModel{
		ID:     "foo-queries-numeric/3",
		Number: 3,
	}
	require.NoError(t, collection.Put(ctx, foo3))

	foo10 := &FooQueryNumericModel{
		ID:     "foo-queries-numeric/10",
		Number: 10,
	}
	require.NoError(t, collection.Put(ctx, foo10))

	require.NoError(t, sess.SaveChanges(ctx))

	var models []*FooQueryNumericModel
	require.NoError(t, collection.OrderBy("-Number").GetAll(ctx, &models))

	require.Len(t, models, 4)
	require.Equal(t, models[0].ID, "foo-queries-numeric/10")
	require.Equal(t, models[1].ID, "foo-queries-numeric/3")
	require.Equal(t, models[2].ID, "foo-queries-numeric/2")
	require.Equal(t, models[3].ID, "foo-queries-numeric/1")
}

func TestQueryWithInclude(t *testing.T) {
	db := initQueryTestbed(t)
	ctx, sess := db.NewSession(context.Background())
	collection := db.Collection(new(FooQueryIncludeModel))

	foo1 := &FooQueryIncludeModel{
		ID:          "foo-queries-include/1",
		DisplayName: "Include1",
		FooQuery:    "foo-queries/1",
	}
	require.NoError(t, collection.Put(ctx, foo1))
	foo2 := &FooQueryIncludeModel{
		ID:          "foo-queries-include/2",
		DisplayName: "Include2",
		FooQuery:    "foo-queries/2",
	}
	require.NoError(t, collection.Put(ctx, foo2))

	require.NoError(t, sess.SaveChanges(ctx))

	var models []*FooQueryIncludeModel
	require.NoError(t, collection.GetAll(ctx, &models, Include("FooQuery")))

	require.Len(t, models, 2)
	{
		model := models[0]
		require.Equal(t, model.ID, "foo-queries-include/1")

		var foo *FooQueryModel
		require.NoError(t, sess.Load(model.FooQuery, &foo))

		require.Equal(t, foo.ID, "foo-queries/1")
	}
	{
		model := models[1]
		require.Equal(t, model.ID, "foo-queries-include/2")

		var foo *FooQueryModel
		require.NoError(t, sess.Load(model.FooQuery, &foo))

		require.Equal(t, foo.ID, "foo-queries/2")
	}
}

type FooQueryIndexModel struct {
	ModelTracking

	ID          string
	DisplayName string
}

func (src *FooQueryIndexModel) Collection() string {
	return "FooQueryIndexModels"
}

type FooProjected struct {
	IndexModel

	ID             string
	DisplayName    string
	AltDisplayName string
}

func TestQueryIndexID(t *testing.T) {
	ctx := context.Background()
	db := initQueryTestbed(t)
	collection := db.Collection(new(FooQueryIndexModel))

	foo := &FooQueryIndexModel{
		ID:          "foo-queries-index/foo1",
		DisplayName: "foo1",
	}
	require.NoError(t, collection.Put(ctx, foo))
	foo = &FooQueryIndexModel{
		ID:          "foo-queries-index/foo2",
		DisplayName: "foo2",
	}
	require.NoError(t, collection.Put(ctx, foo))

	index := Index{
		Maps:  []string{`from doc in docs.FooQueryIndexModels select new { AltDisplayName = doc.DisplayName + "-index" }`},
		Store: []string{"AltDisplayName"},
	}
	require.NoError(t, db.CreateIndex(context.Background(), "QueryIndex", index))

	var results []*FooProjected
	require.NoError(t, db.QueryIndex("QueryIndex", new(FooProjected)).Select("AltDisplayName", "DisplayName").GetAll(ctx, &results))

	require.Len(t, results, 2)

	require.Equal(t, results[0].ID, "foo-queries-index/foo1")
	require.Equal(t, results[0].DisplayName, "foo1")
	require.Equal(t, results[0].AltDisplayName, "foo1-index")

	require.Equal(t, results[1].ID, "foo-queries-index/foo2")
	require.Equal(t, results[1].DisplayName, "foo2")
	require.Equal(t, results[1].AltDisplayName, "foo2-index")
}
