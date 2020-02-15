package database

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIteratorNextCallHooks(t *testing.T) {
	initDatabase(t)
	defer closeDatabase(t)
	ctx := context.Background()

	m := new(testingAutoModel)
	require.Nil(t, testingsAuto.Put(ctx, m))

	m = new(testingAutoModel)
	require.Nil(t, testingsAuto.Put(ctx, m))

	var models []*testingAutoModel
	require.Nil(t, testingsAuto.GetAll(ctx, &models))

	require.Len(t, models, 2)

	require.True(t, models[0].IsInserted())
	require.EqualValues(t, 0, models[0].Tracking().StoredRevision())

	require.True(t, models[1].IsInserted())
	require.EqualValues(t, 0, models[1].Tracking().StoredRevision())
}
