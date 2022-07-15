package pagination

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"

	"libs.altipla.consulting/naming"
	"libs.altipla.consulting/rdb"
)

type RDBModel struct {
	rdb.ModelTracking

	ID string
}

func (model *RDBModel) Collection() string {
	return "RDBModels"
}

func initRDBTestbed(t *testing.T) *rdb.Database {
	ctx := context.Background()

	db, err := rdb.Open("http://localhost:13000", "RDBModels", rdb.WithLocalCreate(), rdb.WithStrongConsistency())
	require.NoError(t, err)
	collection := db.Collection(new(RDBModel))
	require.NoError(t, collection.DeleteEverything(ctx))

	for i := 0; i < 10; i++ {
		model := &RDBModel{ID: naming.Generate("rdbmodels", i)}
		require.NoError(t, collection.Put(ctx, model))
	}

	return db
}

func initCollection(db *rdb.Database) *rdb.Query {
	return db.Collection(new(RDBModel)).Query
}

func TestRDBTokenNextPrevPageTokens(t *testing.T) {
	ctx := context.Background()
	db := initRDBTestbed(t)

	var models []*RDBModel
	var pager *TokenController

	// Page 1 with no token.
	pager = NewRDBToken(initCollection(db), FromToken(2, ""))
	require.NoError(t, pager.Fetch(ctx, &models))
	require.Len(t, models, 2)
	require.Equal(t, models[0].ID, "rdbmodels/0")
	require.Equal(t, models[1].ID, "rdbmodels/1")
	require.Equal(t, pager.NextPageToken(), "plre6r44fy")
	require.Empty(t, pager.PrevPageToken())
	require.EqualValues(t, pager.TotalSize(), 10)

	// Page 2 with next page token.
	pager = NewRDBToken(initCollection(db), FromToken(2, "plre6r44fy"))
	require.NoError(t, pager.Fetch(ctx, &models))
	require.Len(t, models, 2)
	require.Equal(t, models[0].ID, "rdbmodels/2")
	require.Equal(t, models[1].ID, "rdbmodels/3")
	require.Equal(t, pager.NextPageToken(), "krnlen44fz")
	require.Equal(t, pager.PrevPageToken(), "2gexje33fg")

	// Page 3 with next page token.
	pager = NewRDBToken(initCollection(db), FromToken(2, "krnlen44fz"))
	require.NoError(t, pager.Fetch(ctx, &models))
	require.Len(t, models, 2)
	require.Equal(t, models[0].ID, "rdbmodels/4")
	require.Equal(t, models[1].ID, "rdbmodels/5")
	require.Equal(t, pager.NextPageToken(), "4l2op233f4")
	require.Equal(t, pager.PrevPageToken(), "plre6r44fy")

	// Page 1 with prev page token.
	pager = NewRDBToken(initCollection(db), FromToken(2, "2gexje33fg"))
	require.NoError(t, pager.Fetch(ctx, &models))
	require.Len(t, models, 2)
	require.Equal(t, models[0].ID, "rdbmodels/0")
	require.Equal(t, models[1].ID, "rdbmodels/1")
	require.Equal(t, pager.NextPageToken(), "plre6r44fy")
	require.Empty(t, pager.PrevPageToken())
}

func TestRDBTokenLastCount(t *testing.T) {
	ctx := context.Background()
	db := initRDBTestbed(t)

	var models []*RDBModel
	var pager *TokenController

	pager = NewRDBToken(initCollection(db), FromToken(4, ""))
	require.NoError(t, pager.Fetch(ctx, &models))
	require.Len(t, models, 4)
	require.Equal(t, models[0].ID, "rdbmodels/0")
	require.Equal(t, models[1].ID, "rdbmodels/1")
	require.Equal(t, models[2].ID, "rdbmodels/2")
	require.Equal(t, models[3].ID, "rdbmodels/3")
	require.Equal(t, pager.NextPageToken(), "389ozr99ud")

	pager = NewRDBToken(initCollection(db), FromToken(4, "389ozr99ud"))
	require.NoError(t, pager.Fetch(ctx, &models))
	require.Len(t, models, 4)
	require.Equal(t, models[0].ID, "rdbmodels/4")
	require.Equal(t, models[1].ID, "rdbmodels/5")
	require.Equal(t, models[2].ID, "rdbmodels/6")
	require.Equal(t, models[3].ID, "rdbmodels/7")
	require.Equal(t, pager.NextPageToken(), "167ezx77bo")

	pager = NewRDBToken(initCollection(db), FromToken(4, "167ezx77bo"))
	require.NoError(t, pager.Fetch(ctx, &models))
	require.Len(t, models, 2)
	require.Equal(t, models[0].ID, "rdbmodels/8")
	require.Equal(t, models[1].ID, "rdbmodels/9")
	require.Empty(t, pager.NextPageToken())
}

func TestRDBPaged(t *testing.T) {
	ctx := context.Background()
	db := initRDBTestbed(t)

	var models []*RDBModel
	var pager *PagedController

	pager = NewRDBPaged(initCollection(db), FromPaged(4, 1, 0))
	require.NoError(t, pager.Fetch(ctx, &models))
	require.Len(t, models, 4)
	require.Equal(t, models[0].ID, "rdbmodels/0")
	require.Equal(t, models[1].ID, "rdbmodels/1")
	require.Equal(t, models[2].ID, "rdbmodels/2")
	require.Equal(t, models[3].ID, "rdbmodels/3")
	require.EqualValues(t, pager.Checksum(), 304987742)

	pager = NewRDBPaged(initCollection(db), FromPaged(4, 2, 304987742))
	require.NoError(t, pager.Fetch(ctx, &models))
	require.Len(t, models, 4)
	require.Equal(t, models[0].ID, "rdbmodels/4")
	require.Equal(t, models[1].ID, "rdbmodels/5")
	require.Equal(t, models[2].ID, "rdbmodels/6")
	require.Equal(t, models[3].ID, "rdbmodels/7")
	require.EqualValues(t, pager.Checksum(), 304987742)

	pager = NewRDBPaged(initCollection(db), FromPaged(4, 3, 304987742))
	require.NoError(t, pager.Fetch(ctx, &models))
	require.Len(t, models, 2)
	require.Equal(t, models[0].ID, "rdbmodels/8")
	require.Equal(t, models[1].ID, "rdbmodels/9")
	require.EqualValues(t, pager.Checksum(), 304987742)
	require.False(t, pager.OutOfBounds())

	pager = NewRDBPaged(initCollection(db), FromPaged(4, 4, 304987742))
	require.NoError(t, pager.Fetch(ctx, &models))
	require.Empty(t, models)
	require.EqualValues(t, pager.Checksum(), 304987742)
	require.True(t, pager.OutOfBounds())
}

func TestRDBPagedNextPrevURLs(t *testing.T) {
	ctx := context.Background()
	db := initRDBTestbed(t)

	var models []*RDBModel
	var pager *PagedController

	req := httptest.NewRequest(http.MethodGet, "/rdbmodels?page-size=4", nil)
	pager = NewRDBPaged(initCollection(db), FromRequest(req))
	require.NoError(t, pager.Fetch(ctx, &models))
	require.Len(t, models, 4)
	require.Equal(t, models[0].ID, "rdbmodels/0")
	require.Equal(t, models[1].ID, "rdbmodels/1")
	require.Equal(t, models[2].ID, "rdbmodels/2")
	require.Equal(t, models[3].ID, "rdbmodels/3")
	u := &url.URL{Path: "/rdbmodels", RawQuery: "page-size=4"}
	require.Nil(t, pager.PrevPageURL(u))
	u = &url.URL{Path: "/rdbmodels", RawQuery: "page-size=4"}
	require.Equal(t, pager.NextPageURL(u).String(), "/rdbmodels?checksum=304987742&page=2&page-size=4")
	require.Empty(t, pager.PrevPageURLString(req))
	require.Equal(t, pager.NextPageURLString(req), "/rdbmodels?checksum=304987742&page=2&page-size=4")

	req = httptest.NewRequest(http.MethodGet, "/rdbmodels?checksum=304987742&page=2&page-size=4", nil)
	pager = NewRDBPaged(initCollection(db), FromRequest(req))
	require.NoError(t, pager.Fetch(ctx, &models))
	require.Len(t, models, 4)
	require.Equal(t, models[0].ID, "rdbmodels/4")
	require.Equal(t, models[1].ID, "rdbmodels/5")
	require.Equal(t, models[2].ID, "rdbmodels/6")
	require.Equal(t, models[3].ID, "rdbmodels/7")
	u = &url.URL{Path: "/rdbmodels", RawQuery: "page-size=4"}
	require.Equal(t, pager.PrevPageURL(u).String(), "/rdbmodels?page-size=4")
	u = &url.URL{Path: "/rdbmodels", RawQuery: "page-size=4"}
	require.Equal(t, pager.NextPageURL(u).String(), "/rdbmodels?checksum=304987742&page=3&page-size=4")
	require.Equal(t, pager.PrevPageURLString(req), "/rdbmodels?page-size=4")
	require.Equal(t, pager.NextPageURLString(req), "/rdbmodels?checksum=304987742&page=3&page-size=4")

	req = httptest.NewRequest(http.MethodGet, "/rdbmodels?checksum=304987742&page=3&page-size=4", nil)
	pager = NewRDBPaged(initCollection(db), FromRequest(req))
	require.NoError(t, pager.Fetch(ctx, &models))
	require.Len(t, models, 2)
	require.Equal(t, models[0].ID, "rdbmodels/8")
	require.Equal(t, models[1].ID, "rdbmodels/9")
	u = &url.URL{Path: "/rdbmodels", RawQuery: "page-size=4"}
	require.Equal(t, pager.PrevPageURL(u).String(), "/rdbmodels?checksum=304987742&page=2&page-size=4")
	u = &url.URL{Path: "/rdbmodels", RawQuery: "page-size=4"}
	require.Nil(t, pager.NextPageURL(u))
	require.Equal(t, pager.PrevPageURLString(req), "/rdbmodels?checksum=304987742&page=2&page-size=4")
	require.Empty(t, pager.NextPageURLString(req))
}

func TestRDBPagedOutOfBounds(t *testing.T) {
	ctx := context.Background()
	db := initRDBTestbed(t)

	var models []*RDBModel
	var pager *PagedController

	req := httptest.NewRequest(http.MethodGet, "/rdbmodels?checksum=304987742&page=4&page-size=4", nil)
	pager = NewRDBPaged(initCollection(db), FromRequest(req))
	require.NoError(t, pager.Fetch(ctx, &models))
	require.Empty(t, models)
	require.True(t, pager.OutOfBounds())
	u := &url.URL{Path: "/rdbmodels", RawQuery: "page-size=4"}
	require.Equal(t, pager.PrevPageURL(u).String(), "/rdbmodels?checksum=304987742&page=3&page-size=4")
	u = &url.URL{Path: "/rdbmodels", RawQuery: "page-size=4"}
	require.Nil(t, pager.NextPageURL(u))
	require.Equal(t, pager.PrevPageURLString(req), "/rdbmodels?checksum=304987742&page=3&page-size=4")
	require.Empty(t, pager.NextPageURLString(req))

	req = httptest.NewRequest(http.MethodGet, "/rdbmodels?checksum=304987742&page=40&page-size=4", nil)
	pager = NewRDBPaged(initCollection(db), FromRequest(req))
	require.NoError(t, pager.Fetch(ctx, &models))
	require.Empty(t, models)
	require.True(t, pager.OutOfBounds())
	u = &url.URL{Path: "/rdbmodels", RawQuery: "page-size=4"}
	require.Equal(t, pager.PrevPageURL(u).String(), "/rdbmodels?checksum=304987742&page=3&page-size=4")
	u = &url.URL{Path: "/rdbmodels", RawQuery: "page-size=4"}
	require.Nil(t, pager.NextPageURL(u))
	require.Equal(t, pager.PrevPageURLString(req), "/rdbmodels?checksum=304987742&page=3&page-size=4")
	require.Empty(t, pager.NextPageURLString(req))
}
