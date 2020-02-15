package database

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInsertedAfterGet(t *testing.T) {
	initDatabase(t)
	defer closeDatabase(t)

	m := new(testingModel)
	tracking := m.Tracking()

	require.False(t, tracking.IsInserted())

	require.NoError(t, tracking.AfterGet(nil))

	require.True(t, tracking.IsInserted())
}

func TestInsertedAfterPut(t *testing.T) {
	initDatabase(t)
	defer closeDatabase(t)

	m := new(testingModel)
	tracking := m.Tracking()

	require.False(t, tracking.IsInserted())

	require.NoError(t, tracking.AfterPut(nil))

	require.True(t, tracking.IsInserted())
}

func TestInsertedAfterDelete(t *testing.T) {
	initDatabase(t)
	defer closeDatabase(t)

	m := new(testingModel)
	tracking := m.Tracking()
	require.NoError(t, tracking.AfterGet(nil))

	require.True(t, tracking.IsInserted())

	require.NoError(t, tracking.AfterDelete(nil))

	require.False(t, tracking.IsInserted())
}

func TestIncrementRevisionAfterGet(t *testing.T) {
	initDatabase(t)
	defer closeDatabase(t)

	m := new(testingModel)
	tracking := m.Tracking()

	require.EqualValues(t, -1, tracking.StoredRevision())
	require.EqualValues(t, 0, tracking.Revision)

	require.NoError(t, tracking.AfterGet(nil))

	require.EqualValues(t, 0, tracking.StoredRevision())
	require.EqualValues(t, 1, tracking.Revision)
}

func TestIncrementRevisionAfterPut(t *testing.T) {
	initDatabase(t)
	defer closeDatabase(t)

	m := new(testingModel)
	tracking := m.Tracking()

	require.EqualValues(t, -1, tracking.StoredRevision())
	require.EqualValues(t, 0, tracking.Revision)

	require.NoError(t, tracking.AfterPut(nil))

	require.EqualValues(t, 0, tracking.StoredRevision())
	require.EqualValues(t, 1, tracking.Revision)
}
