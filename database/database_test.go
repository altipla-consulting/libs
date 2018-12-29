package database

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	testDB            *Database
	testings          *Collection
	testingsAuto      *Collection
	testingsHooker    *Collection
	testingsRelParent *Collection
	testingsRelChild  *Collection
)

type testingModel struct {
	ModelTracking

	Code    string `db:"code,pk"`
	Name    string `db:"name"`
	Ignored string `db:"-"`
}

func (model *testingModel) TableName() string {
	return "testing"
}

type testingAutoModel struct {
	ModelTracking

	ID      int64  `db:"id,pk"`
	Name    string `db:"name"`
	Ignored string `db:"-"`
}

func (model *testingAutoModel) TableName() string {
	return "testing_auto"
}

type testingHooker struct {
	ModelTracking

	Code     string `db:"code,pk"`
	Executed bool   `db:"executed"`
	Changed  string `db:"changed"`
}

func (model *testingHooker) TableName() string {
	return "testing_hooker"
}

func (model *testingHooker) OnBeforePutHook() error {
	model.Changed = "changed"

	return nil
}

func (model *testingHooker) OnAfterPutHook() error {
	model.Executed = true

	return nil
}

type testingRelParent struct {
	ModelTracking

	ID int64 `db:"id,pk"`
}

func (model *testingRelParent) TableName() string {
	return "testing_relparent"
}

type testingRelChild struct {
	ModelTracking

	ID int64 `db:"id,pk"`

	Parent int64  `db:"parent"`
	Foo    string `db:"foo"`
}

func (model *testingRelChild) TableName() string {
	return "testing_relchild"
}

func initDatabase(t *testing.T) {
	var err error
	testDB, err = Open(Credentials{
		User:      "dev-user",
		Password:  "dev-password",
		Address:   "localhost:3306",
		Database:  "default",
		Charset:   "utf8mb4",
		Collation: "utf8mb4_bin",
	}, WithDebug(os.Getenv("DEBUG") == "true"))
	require.Nil(t, err)

	require.Nil(t, testDB.Exec(`DROP TABLE IF EXISTS testing`))
	err = testDB.Exec(`
    CREATE TABLE testing (
      code VARCHAR(191),
      name VARCHAR(191),
      revision INT(11) NOT NULL,

      PRIMARY KEY(code)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;
  `)
	require.Nil(t, err)

	require.Nil(t, testDB.Exec(`DROP TABLE IF EXISTS testing_auto`))
	err = testDB.Exec(`
    CREATE TABLE testing_auto (
      id INT(11) NOT NULL AUTO_INCREMENT,
      name VARCHAR(191),
      revision INT(11) NOT NULL,

      PRIMARY KEY(id)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;
  `)
	require.Nil(t, err)

	require.Nil(t, testDB.Exec(`DROP TABLE IF EXISTS testing_hooker`))
	err = testDB.Exec(`
    CREATE TABLE testing_hooker (
      code VARCHAR(191),
      executed BOOLEAN,
      changed VARCHAR(191),
      revision INT(11) NOT NULL,

      PRIMARY KEY(code)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;
  `)
	require.Nil(t, err)

	require.Nil(t, testDB.Exec(`DROP TABLE IF EXISTS testing_relparent`))
	err = testDB.Exec(`
    CREATE TABLE testing_relparent (
      id INT(11) NOT NULL AUTO_INCREMENT,
      revision INT(11) NOT NULL,

      PRIMARY KEY(id)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;
  `)
	require.Nil(t, err)

	require.Nil(t, testDB.Exec(`DROP TABLE IF EXISTS testing_relchild`))
	err = testDB.Exec(`
    CREATE TABLE testing_relchild (
      id INT(11) NOT NULL AUTO_INCREMENT,
      parent INT(11),
      foo VARCHAR(191) NOT NULL,
      revision INT(11) NOT NULL,

      PRIMARY KEY(id)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;
  `)
	require.Nil(t, err)

	testings = testDB.Collection(new(testingModel))
	testingsAuto = testDB.Collection(new(testingAutoModel))
	testingsHooker = testDB.Collection(new(testingHooker))
	testingsRelParent = testDB.Collection(new(testingRelParent))
	testingsRelChild = testDB.Collection(new(testingRelChild))
}

func closeDatabase() {
	testDB.Close()
}

func TestQueryRow(t *testing.T) {
	initDatabase(t)

	model := &testingModel{
		Code: "Test",
		Name: "test",
	}
	require.Nil(t, testings.Put(model))

	row := testDB.QueryRow(`SELECT name FROM testing`)

	var name string
	require.NoError(t, row.Scan(&name))
	require.Equal(t, "test", name)
}
