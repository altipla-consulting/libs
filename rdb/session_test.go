package rdb

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadMultipleTimesSameEntityDoesNotFail(t *testing.T) {
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

	var other *FooCollectionModel
	require.NoError(t, collection.Get(ctx, "foo-collections/3", &other, Include("DisplayName")))

	var first *FooCollectionModel
	require.NoError(t, sess.Load("foo-collections/4", &first))
	require.Equal(t, first.ID, "foo-collections/4")
	require.Equal(t, first.DisplayName, "Foo-Dest")

	var repeated *FooCollectionModel
	require.NoError(t, sess.Load("foo-collections/4", &repeated))
	require.Equal(t, repeated.ID, "foo-collections/4")
	require.Equal(t, repeated.DisplayName, "Foo-Dest")
}
