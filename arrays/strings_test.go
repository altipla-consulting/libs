package arrays

import (
	"testing"

	"github.com/stretchr/testify/require"
	"upper.io/db.v3"
	"upper.io/db.v3/lib/sqlbuilder"
	"upper.io/db.v3/mysql"
)

var (
	stringsSess   sqlbuilder.Database
	stringsModels db.Collection
)

type stringsModel struct {
	ID  int64   `db:"id,omitempty"`
	Foo Strings `db:"foo"`
}

func initStringsDB(t *testing.T) {
	cnf := &mysql.ConnectionURL{
		User:     "dev-user",
		Password: "dev-password",
		Host:     "database:3306",
		Database: "default",
		Options: map[string]string{
			"charset":   "utf8mb4",
			"collation": "utf8mb4_bin",
			"parseTime": "true",
		},
	}
	var err error
	stringsSess, err = mysql.Open(cnf)
	require.Nil(t, err)

	_, err = stringsSess.Exec(`DROP TABLE IF EXISTS arrays_test`)
	require.Nil(t, err)

	_, err = stringsSess.Exec(`
    CREATE TABLE arrays_test (
      id INT(11) NOT NULL AUTO_INCREMENT,
      foo JSON,

      PRIMARY KEY(id)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;
  `)
	require.Nil(t, err)

	stringsModels = stringsSess.Collection("arrays_test")

	require.Nil(t, stringsModels.Truncate())
}

func finishStringsDB() {
	stringsSess.Close()
}

func TestLoadNilStrings(t *testing.T) {
	initStringsDB(t)
	defer finishStringsDB()

	_, err := stringsSess.Exec(`INSERT INTO arrays_test() VALUES ()`)
	require.NoError(t, err)

	model := new(stringsModel)
	require.Nil(t, stringsModels.Find(1).One(model))
}

func TestLoadSaveStrings(t *testing.T) {
	initStringsDB(t)
	defer finishStringsDB()

	model := new(stringsModel)
	require.Nil(t, stringsModels.InsertReturning(model))

	require.EqualValues(t, model.ID, 1)

	other := new(stringsModel)
	require.Nil(t, stringsModels.Find(1).One(other))
}

func TestLoadSaveStringsWithContent(t *testing.T) {
	initStringsDB(t)
	defer finishStringsDB()

	model := &stringsModel{
		Foo: []string{"foo", "bar"},
	}
	require.Nil(t, stringsModels.InsertReturning(model))

	other := new(stringsModel)
	require.Nil(t, stringsModels.Find(1).One(other))

	require.Equal(t, other.Foo, Strings{"foo", "bar"})
	require.EqualValues(t, other.Foo, []string{"foo", "bar"})
}

func TestStringsSearch(t *testing.T) {
	initStringsDB(t)
	defer finishStringsDB()

	model := &stringsModel{
		Foo: []string{"foo", "bar"},
	}
	require.Nil(t, stringsModels.InsertReturning(model))

	model = &stringsModel{
		Foo: []string{"foo", "baz"},
	}
	require.Nil(t, stringsModels.InsertReturning(model))

	models := []*stringsModel{}
	require.Nil(t, stringsModels.Find(SearchStrings("foo"), "foo").All(&models))
	require.Len(t, models, 2)

	require.Nil(t, stringsModels.Find(SearchStrings("foo"), "bar").All(&models))
	require.Len(t, models, 1)
	require.EqualValues(t, models[0].ID, 1)
}

func TestStringsSaveNil(t *testing.T) {
	initStringsDB(t)
	defer finishStringsDB()

	model := new(stringsModel)
	require.Nil(t, stringsModels.InsertReturning(model))

	row, err := stringsSess.QueryRow(`SELECT foo FROM arrays_test`)
	require.NoError(t, err)

	var foo string
	require.NoError(t, row.Scan(&foo))
	require.Equal(t, "[]", foo)
}
