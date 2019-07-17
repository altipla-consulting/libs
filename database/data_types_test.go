package database

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type testingModelNullString struct {
	ModelTracking

	Code string         `db:"code,pk"`
	Name NullableString `db:"name"`
}

func (model *testingModelNullString) TableName() string {
	return "testing"
}

func TestNullableStringDefaultsValid(t *testing.T) {
	initDatabase(t)
	defer closeDatabase()

	c := testDB.Collection(new(testingModelNullString))

	m := &testingModelNullString{
		Code: "foo",
	}
	require.NoError(t, c.Put(m))

	other := &testingModel{
		Code: "foo",
	}
	require.NoError(t, testings.Get(other))

	require.Empty(t, other.Name)
}

func TestNullableStringStoresValue(t *testing.T) {
	initDatabase(t)
	defer closeDatabase()

	c := testDB.Collection(new(testingModelNullString))

	m := &testingModelNullString{
		Code: "foo",
		Name: "foo value",
	}
	require.NoError(t, c.Put(m))

	other := &testingModel{
		Code: "foo",
	}
	require.NoError(t, testings.Get(other))

	require.Equal(t, other.Name, "foo value")
}
