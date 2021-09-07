package rdb

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"libs.altipla.consulting/datetime"
)

type FooDateTimeModel struct {
	ModelTracking

	ID    string `json:",omitempty"`
	Field DateTime
}

func (model *FooDateTimeModel) Collection() string {
	return "FooDateTimeModels"
}

func initTypesTesbed(t *testing.T) *Database {
	ctx := context.Background()
	db := initTestbed(t)

	require.NoError(t, db.Collection(new(FooDateTimeModel)).DeleteEverything(ctx))
	require.NoError(t, db.Collection(new(FooDateModel)).DeleteEverything(ctx))

	return db
}

func TestDateTimeMarshal(t *testing.T) {
	model := &FooDateTimeModel{
		Field: NewDateTime(time.Date(2006, time.January, 2, 3, 4, 5, 6, time.UTC)),
	}
	data, err := json.Marshal(model)
	require.NoError(t, err)
	require.Equal(t, string(data), `{"Field":"2006-01-02T03:04:05.0000000Z"}`)
}

func TestDateTimeMarshalZero(t *testing.T) {
	model := new(FooDateTimeModel)
	data, err := json.Marshal(model)
	require.NoError(t, err)
	require.Equal(t, string(data), `{"Field":null}`)
}

func TestDateTimeMarshalTimezone(t *testing.T) {
	model := &FooDateTimeModel{
		Field: NewDateTime(time.Date(2006, time.January, 2, 3, 4, 5, 6, datetime.EuropeMadrid())),
	}
	data, err := json.Marshal(model)
	require.NoError(t, err)
	require.Equal(t, string(data), `{"Field":"2006-01-02T02:04:05.0000000Z"}`)
}

func TestDateTimeUnmarshal(t *testing.T) {
	model := new(FooDateTimeModel)
	require.NoError(t, json.Unmarshal([]byte(`{"Field":"2006-01-02T03:04:05.0000000Z"}`), model))

	require.WithinDuration(t, model.Field.Time, time.Date(2006, time.January, 2, 3, 4, 5, 6, time.UTC), 1*time.Second)
}

func TestDateTimeUnmarshalZero(t *testing.T) {
	model := new(FooDateTimeModel)
	require.NoError(t, json.Unmarshal([]byte(`{"Field":null}`), model))

	require.True(t, model.Field.IsZero())
}

func TestDateTimeNullTimeFromNullToFilledNoPanic(t *testing.T) {
	db := initTypesTesbed(t)
	collection := db.Collection(new(FooDateTimeModel))

	foo := &FooDateTimeModel{
		ID: "foo-date-time/3",
	}
	require.NoError(t, collection.Put(context.Background(), foo))

	var other *FooDateTimeModel
	require.NoError(t, collection.Get(context.Background(), "foo-date-time/3", &other))
	other.Field = NewDateTime(time.Now())
	require.NoError(t, collection.Put(context.Background(), other))
}

type FooDateModel struct {
	ModelTracking

	ID    string `json:",omitempty"`
	Field Date
}

func (model *FooDateModel) Collection() string {
	return "FooDateModels"
}

func TestDateMarshal(t *testing.T) {
	model := &FooDateModel{
		Field: NewDate(time.Date(2006, time.January, 2, 3, 4, 5, 6, time.UTC)),
	}
	data, err := json.Marshal(model)
	require.NoError(t, err)
	require.Equal(t, string(data), `{"Field":"2006-01-02T00:00:00.0000000Z"}`)
}

func TestDateUnmarshal(t *testing.T) {
	model := new(FooDateModel)
	require.NoError(t, json.Unmarshal([]byte(`{"Field":"2006-01-02T00:00:00.0000000Z"}`), model))

	require.WithinDuration(t, model.Field.Time, time.Date(2006, time.January, 2, 0, 0, 0, 0, time.UTC), 1*time.Second)
}

func TestDateUnmarshalObject(t *testing.T) {
	model := new(FooDateModel)
	require.NoError(t, json.Unmarshal([]byte(`{"Field":{"day":2,"month":1,"year":2006}}`), model))

	require.WithinDuration(t, model.Field.Time, time.Date(2006, time.January, 2, 0, 0, 0, 0, time.UTC), 1*time.Second)
}
