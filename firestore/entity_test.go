package firestore

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

type entityFake struct {
	Name string
	Foo  string
}

func (fake *entityFake) Collection() string {
	return "entity_fakes"
}

func (fake *entityFake) Key() string {
	return fake.Name
}

func TestEntityCyclePutGet(t *testing.T) {
	db := initDatabase(t)
	ctx := context.Background()
	c := db.Entity(new(entityFake))

	fake := &entityFake{
		Name: "foo-name",
		Foo:  "foo-value",
	}
	require.NoError(t, c.Put(ctx, fake))

	other := &entityFake{
		Name: "foo-name",
	}
	require.NoError(t, c.Get(ctx, other))

	require.Equal(t, other.Foo, "foo-value")
}

func TestEntityCycleDelete(t *testing.T) {
	db := initDatabase(t)
	ctx := context.Background()
	c := db.Entity(new(entityFake))

	fake := &entityFake{
		Name: "foo-name",
		Foo:  "foo-value",
	}
	require.NoError(t, c.Put(ctx, fake))

	require.NoError(t, c.Delete(ctx, fake))

	require.EqualError(t, c.Get(ctx, fake), ErrNoSuchEntity.Error())
}
