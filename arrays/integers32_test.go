package arrays

import (
	"testing"

	"github.com/stretchr/testify/require"
	"upper.io/db.v3"
	"upper.io/db.v3/lib/sqlbuilder"
	"upper.io/db.v3/mysql"
)

var (
	integers32Sess   sqlbuilder.Database
	integers32Models db.Collection
)

type integers32Model struct {
	ID  int32      `db:"id,omitempty"`
	Foo Integers32 `db:"foo"`
}

func initIntegers32DB(t *testing.T) {
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
	integers32Sess, err = mysql.Open(cnf)
	require.Nil(t, err)

	_, err = integers32Sess.Exec(`DROP TABLE IF EXISTS arrays_test`)
	require.Nil(t, err)

	_, err = integers32Sess.Exec(`
    CREATE TABLE arrays_test (
      id INT(11) NOT NULL AUTO_INCREMENT,
      foo JSON,

      PRIMARY KEY(id)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;
  `)
	require.Nil(t, err)

	integers32Models = integers32Sess.Collection("arrays_test")

	require.Nil(t, integers32Models.Truncate())
}

func finishIntegers32DB() {
	integers32Sess.Close()
}

func TestLoadNilIntegers32(t *testing.T) {
	initIntegers32DB(t)
	defer finishIntegers32DB()

	_, err := integers32Sess.Exec(`INSERT INTO arrays_test() VALUES ()`)
	require.NoError(t, err)

	model := new(integers32Model)
	require.Nil(t, integers32Models.Find(1).One(model))
}

func TestLoadSaveIntegers32(t *testing.T) {
	initIntegers32DB(t)
	defer finishIntegers32DB()

	model := new(integers32Model)
	require.Nil(t, integers32Models.InsertReturning(model))

	require.EqualValues(t, model.ID, 1)

	other := new(integers32Model)
	require.Nil(t, integers32Models.Find(1).One(other))
}

func TestLoadSaveIntegers32WithContent(t *testing.T) {
	initIntegers32DB(t)
	defer finishIntegers32DB()

	model := &integers32Model{
		Foo: []int32{3, 4},
	}
	require.Nil(t, integers32Models.InsertReturning(model))

	other := new(integers32Model)
	require.Nil(t, integers32Models.Find(1).One(other))

	require.Equal(t, other.Foo, Integers32{3, 4})
	require.EqualValues(t, other.Foo, []int32{3, 4})
}

func TestIntegers32Search(t *testing.T) {
	initIntegers32DB(t)
	defer finishIntegers32DB()

	model := &integers32Model{
		Foo: []int32{10, 20},
	}
	require.Nil(t, integers32Models.InsertReturning(model))

	model = &integers32Model{
		Foo: []int32{10, 30},
	}
	require.Nil(t, integers32Models.InsertReturning(model))

	models := []*integers32Model{}

	require.Nil(t, integers32Models.Find(SearchIntegers32("foo"), 10).All(&models))
	require.Len(t, models, 2)

	require.Nil(t, integers32Models.Find(SearchIntegers32("foo"), "10").All(&models))
	require.Len(t, models, 2)

	require.Nil(t, integers32Models.Find(SearchIntegers32("foo"), 20).All(&models))
	require.Len(t, models, 1)
	require.EqualValues(t, models[0].ID, 1)
}

func TestIntegers32SaveNil(t *testing.T) {
	initIntegers32DB(t)
	defer finishIntegers32DB()

	model := new(integers32Model)
	require.Nil(t, integers32Models.InsertReturning(model))

	row, err := integers32Sess.QueryRow(`SELECT foo FROM arrays_test`)
	require.NoError(t, err)

	var foo string
	require.NoError(t, row.Scan(&foo))
	require.Equal(t, "[]", foo)
}
