package database

import (
	"context"
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
	ctx := context.Background()

	var err error
	testDB, err = Open(Credentials{
		User:      "dev-user",
		Password:  "dev-password",
		Address:   "database:3306",
		Database:  "default",
		Charset:   "utf8mb4",
		Collation: "utf8mb4_bin",
	}, WithDebug(os.Getenv("DEBUG") == "true"))
	require.Nil(t, err)

	require.Nil(t, testDB.Exec(ctx, `DROP TABLE IF EXISTS testing`))
	err = testDB.Exec(ctx, `
    CREATE TABLE testing (
      code VARCHAR(191),
      name VARCHAR(191),
      revision INT(11) NOT NULL,

      PRIMARY KEY(code)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;
  `)
	require.Nil(t, err)

	require.Nil(t, testDB.Exec(ctx, `DROP TABLE IF EXISTS testing_auto`))
	err = testDB.Exec(ctx, `
    CREATE TABLE testing_auto (
      id INT(11) NOT NULL AUTO_INCREMENT,
      name VARCHAR(191),
      revision INT(11) NOT NULL,

      PRIMARY KEY(id)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;
  `)
	require.Nil(t, err)

	require.Nil(t, testDB.Exec(ctx, `DROP TABLE IF EXISTS testing_hooker`))
	err = testDB.Exec(ctx, `
    CREATE TABLE testing_hooker (
      code VARCHAR(191),
      executed BOOLEAN,
      changed VARCHAR(191),
      revision INT(11) NOT NULL,

      PRIMARY KEY(code)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;
  `)
	require.Nil(t, err)

	require.Nil(t, testDB.Exec(ctx, `DROP TABLE IF EXISTS testing_relparent`))
	err = testDB.Exec(ctx, `
    CREATE TABLE testing_relparent (
      id INT(11) NOT NULL AUTO_INCREMENT,
      revision INT(11) NOT NULL,

      PRIMARY KEY(id)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;
  `)
	require.Nil(t, err)

	require.Nil(t, testDB.Exec(ctx, `DROP TABLE IF EXISTS testing_relchild`))
	err = testDB.Exec(ctx, `
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

func closeDatabase(t *testing.T) {
	require.NoError(t, testDB.Close())
}

func TestQueryRow(t *testing.T) {
	initDatabase(t)
	defer closeDatabase(t)

	model := &testingModel{
		Code: "Test",
		Name: "test",
	}
	require.Nil(t, testings.Put(context.Background(), model))

	row := testDB.QueryRow(context.Background(), `SELECT name FROM testing`)

	var name string
	require.NoError(t, row.Scan(&name))
	require.Equal(t, "test", name)
}

type testingModelSelect struct {
	Code string `db:"code"`
	Name string `db:"name"`
}

func TestSelect(t *testing.T) {
	initDatabase(t)
	defer closeDatabase(t)

	model := &testingModel{
		Code: "foo",
		Name: "Foo Name",
	}
	require.Nil(t, testings.Put(context.Background(), model))

	other := new(testingModelSelect)
	require.NoError(t, testDB.Select(context.Background(), other, `SELECT code, name FROM testing`))

	require.Equal(t, other.Code, "foo")
	require.Equal(t, other.Name, "Foo Name")
}

func TestSelectAll(t *testing.T) {
	initDatabase(t)
	defer closeDatabase(t)

	model := &testingModel{
		Code: "foo",
		Name: "Foo Name",
	}
	require.Nil(t, testings.Put(context.Background(), model))
	model = &testingModel{
		Code: "bar",
		Name: "Bar Name",
	}
	require.Nil(t, testings.Put(context.Background(), model))

	var other []*testingModelSelect
	require.NoError(t, testDB.SelectAll(context.Background(), &other, `SELECT code, name FROM testing`))

	require.Len(t, other, 2)
	require.Equal(t, other[0].Code, "bar")
	require.Equal(t, other[0].Name, "Bar Name")
	require.Equal(t, other[1].Code, "foo")
	require.Equal(t, other[1].Name, "Foo Name")
}
