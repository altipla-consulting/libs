package pagination

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"libs.altipla.consulting/database"
)

var (
	testDB   *database.Database
	testings *database.Collection
)

type testingModel struct {
	database.ModelTracking

	Code string `db:"code,pk"`
}

func (model *testingModel) TableName() string {
	return "testing"
}

func initDatabase(t *testing.T) {
	var err error
	testDB, err = database.Open(database.Credentials{
		User:      "dev-user",
		Password:  "dev-password",
		Address:   "database:3306",
		Database:  "default",
		Charset:   "utf8mb4",
		Collation: "utf8mb4_bin",
	}, database.WithDebug(os.Getenv("DEBUG") == "true"))
	require.Nil(t, err)

	require.Nil(t, testDB.Exec(`DROP TABLE IF EXISTS testing`))
	err = testDB.Exec(`
    CREATE TABLE testing (
      code VARCHAR(191),
      revision INT(11) NOT NULL,

      PRIMARY KEY(code)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;
  `)
	require.Nil(t, err)

	testings = testDB.Collection(new(testingModel))
}

func closeDatabase() {
	testDB.Close()
}

func TestMovingBetweenAllPages(t *testing.T) {
	initDatabase(t)
	defer closeDatabase()

	for i := 0; i < 8; i++ {
		foo := &testingModel{
			Code: fmt.Sprintf("foo-%d", i),
		}
		require.NoError(t, testings.Put(foo))
	}

	p := NewPager(testings)
	{
		p.SetInputs("", 3)
		var page []*testingModel
		require.NoError(t, p.Fetch(&page))

		require.Len(t, page, 3)

		require.Equal(t, page[0].Code, "foo-0")
		require.Equal(t, page[1].Code, "foo-1")
		require.Equal(t, page[2].Code, "foo-2")

		require.EqualValues(t, p.TotalSize, 8)
		require.Empty(t, p.PrevPageToken)
		require.NotEmpty(t, p.NextPageToken)
	}
	{
		p.SetInputs(p.NextPageToken, 3)
		var page []*testingModel
		require.NoError(t, p.Fetch(&page))

		require.Len(t, page, 3)

		require.Equal(t, page[0].Code, "foo-3")
		require.Equal(t, page[1].Code, "foo-4")
		require.Equal(t, page[2].Code, "foo-5")

		require.EqualValues(t, p.TotalSize, 8)
		require.NotEmpty(t, p.NextPageToken)
	}
	{
		p.SetInputs(p.NextPageToken, 3)
		var page []*testingModel
		require.NoError(t, p.Fetch(&page))

		require.Len(t, page, 2)

		require.Equal(t, page[0].Code, "foo-6")
		require.Equal(t, page[1].Code, "foo-7")

		require.EqualValues(t, p.TotalSize, 8)
		require.NotEmpty(t, p.PrevPageToken)
		require.Empty(t, p.NextPageToken)
	}
	{
		p.SetInputs(p.PrevPageToken, 3)
		var page []*testingModel
		require.NoError(t, p.Fetch(&page))

		require.Len(t, page, 3)

		require.Equal(t, page[0].Code, "foo-3")
		require.Equal(t, page[1].Code, "foo-4")
		require.Equal(t, page[2].Code, "foo-5")

		require.EqualValues(t, p.TotalSize, 8)
		require.NotEmpty(t, p.NextPageToken)
	}
	{
		p.SetInputs(p.PrevPageToken, 3)
		var page []*testingModel
		require.NoError(t, p.Fetch(&page))

		require.Len(t, page, 3)

		require.Equal(t, page[0].Code, "foo-0")
		require.Equal(t, page[1].Code, "foo-1")
		require.Equal(t, page[2].Code, "foo-2")

		require.EqualValues(t, p.TotalSize, 8)
		require.Empty(t, p.PrevPageToken)
		require.NotEmpty(t, p.NextPageToken)
	}
}

func TestEmptyResultSet(t *testing.T) {
	initDatabase(t)
	defer closeDatabase()

	p := NewPager(testings)
	p.SetInputs("", 3)
	var page []*testingModel
	require.NoError(t, p.Fetch(&page))

	require.Empty(t, page)
}
