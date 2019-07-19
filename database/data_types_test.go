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

type testingModelText struct {
	ModelTracking

	Code string         `db:"code,pk"`
	Name NullableString `db:"name"`
}

func (model *testingModelText) TableName() string {
	return "testing_text"
}

func TestNullableStringText(t *testing.T) {
	initDatabase(t)
	defer closeDatabase()

	require.Nil(t, testDB.Exec(`DROP TABLE IF EXISTS testing_text`))
	err := testDB.Exec(`
    CREATE TABLE testing_text (
      code VARCHAR(191),
      name TEXT,
      revision INT(11) NOT NULL,

      PRIMARY KEY(code)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;
  `)
	require.Nil(t, err)

	c := testDB.Collection(new(testingModelText))

	m := &testingModelText{
		Code: "foo",
	}
	require.NoError(t, c.Put(m))

	other := &testingModelText{
		Code: "foo",
	}
	require.NoError(t, c.Get(other))

	require.Empty(t, other.Name)
}
