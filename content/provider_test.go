package content

import (
	"testing"

	"github.com/stretchr/testify/require"
	"upper.io/db.v3"
	"upper.io/db.v3/lib/sqlbuilder"
	"upper.io/db.v3/mysql"
)

var (
	providerSess   sqlbuilder.Database
	providerModels db.Collection
)

type testProviderModel struct {
	ID          int64    `db:"id,omitempty"`
	Name        Provider `db:"name"`
	Description Provider `db:"description"`
}

func initProviderDB(t *testing.T) {
	cnf := &mysql.ConnectionURL{
		User:     "dev-user",
		Password: "dev-password",
		Host:     "localhost",
		Database: "default",
		Options: map[string]string{
			"charset":   "utf8mb4",
			"collation": "utf8mb4_bin",
			"parseTime": "true",
		},
	}
	var err error
	providerSess, err = mysql.Open(cnf)
	require.Nil(t, err)

	_, err = providerSess.Exec(`DROP TABLE IF EXISTS provider_test`)
	require.Nil(t, err)

	_, err = providerSess.Exec(`
    CREATE TABLE provider_test (
      id INT(11) NOT NULL AUTO_INCREMENT,
      name JSON,
      description JSON,

      PRIMARY KEY(id)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;
  `)
	require.Nil(t, err)

	providerModels = providerSess.Collection("provider_test")

	require.Nil(t, providerModels.Truncate())
}

func finishProviderDB() {
	providerSess.Close()
}

func TestLoadSaveProvider(t *testing.T) {
	initProviderDB(t)
	defer finishProviderDB()

	model := new(testProviderModel)
	require.Nil(t, providerModels.InsertReturning(model))

	require.EqualValues(t, model.ID, 1)

	other := new(testProviderModel)
	require.Nil(t, providerModels.Find(1).One(other))
}

func TestLoadSaveProviderWithContent(t *testing.T) {
	initProviderDB(t)
	defer finishProviderDB()

	model := &testProviderModel{
		Name: Provider{"altipla": "foo", "hotelbeds": "bar"},
	}
	require.Nil(t, providerModels.InsertReturning(model))

	other := new(testProviderModel)
	require.Nil(t, providerModels.Find(1).One(other))

	require.Equal(t, other.Name["altipla"], "foo")
	require.Equal(t, other.Name["hotelbeds"], "bar")
}

func TestProviderGlobalChain(t *testing.T) {
	SetGlobalProviderChain([]string{"altipla", "hotelbeds", "dingus"})

	tests := []struct {
		content Provider
		chain   string
	}{
		{
			Provider{"altipla": "foo", "hotelbeds": "bar", "dingus": "baz"},
			"foo",
		},
		{
			Provider{"other1": "foo", "hotelbeds": "bar", "dingus": "baz"},
			"bar",
		},
		{
			Provider{"other1": "foo", "other2": "bar", "dingus": "baz"},
			"baz",
		},
		{
			Provider{"foo": "bar"},
			"",
		},
	}
	for _, test := range tests {
		require.Equal(t, test.content.Chain(), test.chain)
	}
}

func TestProviderCustomChain(t *testing.T) {
	SetGlobalProviderChain([]string{"dingus", "tirpadvisor", "other"})

	content := Provider{"altipla": "foo", "hotelbeds": "bar", "dingus": "baz"}
	require.Equal(t, content.CustomChain([]string{"altipla", "hotelbeds", "dingus"}), "foo")
}

func TestGlobalChainProvider(t *testing.T) {
	SetGlobalProviderChain([]string{"altipla", "hotelbeds", "dingus"})

	tests := []struct {
		content  Provider
		provider string
	}{
		{
			Provider{"altipla": "foo", "hotelbeds": "bar", "dingus": "baz"},
			"altipla",
		},
		{
			Provider{"other1": "foo", "hotelbeds": "bar", "dingus": "baz"},
			"hotelbeds",
		},
		{
			Provider{"other1": "foo", "other2": "bar", "dingus": "baz"},
			"dingus",
		},
		{
			Provider{"foo": "bar"},
			"",
		},
	}
	for _, test := range tests {
		require.Equal(t, test.content.ChainProvider(), test.provider)
	}
}

func TestCustomChainProvider(t *testing.T) {
	SetGlobalProviderChain([]string{"dingus", "tirpadvisor", "other"})

	content := Provider{"altipla": "foo", "hotelbeds": "bar", "dingus": "baz"}
	require.Equal(t, content.CustomChainProvider([]string{"altipla", "hotelbeds", "dingus"}), "altipla")
}

func TestProviderSaveNil(t *testing.T) {
	initProviderDB(t)
	defer finishProviderDB()

	model := new(testProviderModel)
	require.Nil(t, providerModels.InsertReturning(model))

	row, err := providerSess.QueryRow(`SELECT name FROM provider_test`)
	require.NoError(t, err)

	var name string
	require.NoError(t, row.Scan(&name))
	require.Equal(t, "{}", name)
}

func TestProviderLoadNil(t *testing.T) {
	initProviderDB(t)
	defer finishProviderDB()

	_, err := providerSess.Exec(`INSERT INTO provider_test(name, description) VALUES ('null', 'null')`)
	require.NoError(t, err)

	model := new(testProviderModel)
	require.Nil(t, providerModels.Find(1).One(model))

	require.NotNil(t, model.Name)
	require.Len(t, model.Name, 0)
}
