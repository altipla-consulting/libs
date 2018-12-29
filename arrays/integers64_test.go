package arrays

import (
	"testing"

	"github.com/stretchr/testify/require"
	"upper.io/db.v3"
	"upper.io/db.v3/lib/sqlbuilder"
	"upper.io/db.v3/mysql"
)

var (
	integers64Sess   sqlbuilder.Database
	integers64Models db.Collection
)

type integers64Model struct {
	ID  int64      `db:"id,omitempty"`
	Foo Integers64 `db:"foo"`
}

func initIntegers64DB(t *testing.T) {
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
	integers64Sess, err = mysql.Open(cnf)
	require.Nil(t, err)

	_, err = integers64Sess.Exec(`DROP TABLE IF EXISTS arrays_test`)
	require.Nil(t, err)

	_, err = integers64Sess.Exec(`
    CREATE TABLE arrays_test (
      id INT(11) NOT NULL AUTO_INCREMENT,
      foo JSON,

      PRIMARY KEY(id)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;
  `)
	require.Nil(t, err)

	integers64Models = integers64Sess.Collection("arrays_test")

	require.Nil(t, integers64Models.Truncate())
}

func finishIntegers64DB() {
	integers64Sess.Close()
}

func TestLoadNilIntegers64(t *testing.T) {
	initIntegers64DB(t)
	defer finishIntegers64DB()

	_, err := integers64Sess.Exec(`INSERT INTO arrays_test() VALUES ()`)
	require.NoError(t, err)

	model := new(integers64Model)
	require.Nil(t, integers64Models.Find(1).One(model))
}

func TestLoadSaveIntegers64(t *testing.T) {
	initIntegers64DB(t)
	defer finishIntegers64DB()

	model := new(integers64Model)
	require.Nil(t, integers64Models.InsertReturning(model))

	require.EqualValues(t, model.ID, 1)

	other := new(integers64Model)
	require.Nil(t, integers64Models.Find(1).One(other))
}

func TestLoadSaveIntegers64WithContent(t *testing.T) {
	initIntegers64DB(t)
	defer finishIntegers64DB()

	model := &integers64Model{
		Foo: []int64{3, 4},
	}
	require.Nil(t, integers64Models.InsertReturning(model))

	other := new(integers64Model)
	require.Nil(t, integers64Models.Find(1).One(other))

	require.Equal(t, other.Foo, Integers64{3, 4})
	require.EqualValues(t, other.Foo, []int64{3, 4})
}

func TestIntegers64Search(t *testing.T) {
	initIntegers64DB(t)
	defer finishIntegers64DB()

	model := &integers64Model{
		Foo: []int64{10, 20},
	}
	require.Nil(t, integers64Models.InsertReturning(model))

	model = &integers64Model{
		Foo: []int64{10, 30},
	}
	require.Nil(t, integers64Models.InsertReturning(model))

	models := []*integers64Model{}

	require.Nil(t, integers64Models.Find(SearchIntegers64("foo"), 10).All(&models))
	require.Len(t, models, 2)

	require.Nil(t, integers64Models.Find(SearchIntegers64("foo"), "10").All(&models))
	require.Len(t, models, 2)

	require.Nil(t, integers64Models.Find(SearchIntegers64("foo"), 20).All(&models))
	require.Len(t, models, 1)
	require.EqualValues(t, models[0].ID, 1)
}

func TestIntegers64SaveNil(t *testing.T) {
	initIntegers64DB(t)
	defer finishIntegers64DB()

	model := new(integers64Model)
	require.Nil(t, integers64Models.InsertReturning(model))

	row, err := integers64Sess.QueryRow(`SELECT foo FROM arrays_test`)
	require.NoError(t, err)

	var foo string
	require.NoError(t, row.Scan(&foo))
	require.Equal(t, "[]", foo)
}
