package database

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWithAfterPut(t *testing.T) {
	initDatabase(t)
	defer closeDatabase()

	var called int
	fn := func(ctx context.Context, instance Model) error {
		called++
		return nil
	}
	c := testDB.Collection(new(testingModel), WithAfterPut(fn))

	m := &testingModel{
		Code: "foo",
		Name: "bar",
	}
	require.Nil(t, c.Put(m))

	require.Equal(t, called, 1)
}
