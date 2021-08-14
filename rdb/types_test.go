package rdb

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"libs.altipla.consulting/datetime"
)

type FooNullModel struct {
	ModelTracking

	ID    string `json:",omitempty"`
	Field DateTime
}

func (model *FooNullModel) Collection() string {
	return "FooNullModels"
}

func TestTimeMarshal(t *testing.T) {
	model := &FooNullModel{
		Field: NewDateTime(time.Date(2006, time.January, 2, 3, 4, 5, 6, time.UTC)),
	}
	data, err := json.Marshal(model)
	require.NoError(t, err)
	require.Equal(t, string(data), `{"Field":"2006-01-02T03:04:05.0000000Z"}`)
}

func TestTimeMarshalZero(t *testing.T) {
	model := new(FooNullModel)
	data, err := json.Marshal(model)
	require.NoError(t, err)
	require.Equal(t, string(data), `{"Field":null}`)
}

func TestTimeMarshalTimezone(t *testing.T) {
	model := &FooNullModel{
		Field: NewDateTime(time.Date(2006, time.January, 2, 3, 4, 5, 6, datetime.EuropeMadrid())),
	}
	data, err := json.Marshal(model)
	require.NoError(t, err)
	require.Equal(t, string(data), `{"Field":"2006-01-02T02:04:05.0000000Z"}`)
}

func TestTimeUnmarshal(t *testing.T) {
	model := new(FooNullModel)
	require.NoError(t, json.Unmarshal([]byte(`{"Field":"2006-01-02T03:04:05.0000000Z"}`), model))

	require.WithinDuration(t, model.Field.Time, time.Date(2006, time.January, 2, 3, 4, 5, 6, time.UTC), 1*time.Second)
}

func TestTimeUnmarshalZero(t *testing.T) {
	model := new(FooNullModel)
	require.NoError(t, json.Unmarshal([]byte(`{"Field":null}`), model))

	require.True(t, model.Field.IsZero())
}

type NullTimeModel struct {
	ModelTracking

	ID  string
	Foo DateTime
}

func (model *NullTimeModel) Collection() string {
	return "NullTimeModels"
}

func TestTimeNullTimeFromNullToFilledNoPanic(t *testing.T) {
	db := initCollectionTestbed(t)
	ctx, sess := db.NewSession(context.Background())
	collection := db.Collection(new(FooCollectionModel))

	foo := &NullTimeModel{
		ID:  "foo-null-time/3",
		Foo: NewDateTime(time.Time{}),
	}
	require.NoError(t, collection.Put(ctx, foo))
	require.NoError(t, sess.SaveChanges(ctx))

	ctx, sess = db.NewSession(context.Background())
	var other *NullTimeModel
	require.NoError(t, collection.Get(ctx, "foo-null-time/3", &other))

	other.Foo = NewDateTime(time.Now())

	require.NoError(t, collection.Put(ctx, other))
	require.NoError(t, sess.SaveChanges(ctx))
}
