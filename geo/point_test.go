package geo

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/require"

	_ "github.com/go-sql-driver/mysql"
)

var (
	pointSess *sql.DB
)

func initPointDB(t *testing.T) {
	var err error
	pointSess, err = sql.Open("mysql", "dev-user:dev-password@tcp(localhost:3306)/default?parseTime=true&charset=utf8mb4&collation=utf8mb4_bin")
	require.Nil(t, err)

	_, err = pointSess.Exec(`DROP TABLE IF EXISTS points`)
	require.Nil(t, err)

	_, err = pointSess.Exec(`
    CREATE TABLE points (
      id INT(11) NOT NULL AUTO_INCREMENT,
      location POINT,

      PRIMARY KEY(id)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;
  `)
	require.Nil(t, err)
}

func finishPointDB() {
	pointSess.Close()
}

func TestLoadSavePoint(t *testing.T) {
	initPointDB(t)
	defer finishPointDB()

	result, err := pointSess.Exec(`INSERT INTO points(location) VALUES (?)`, Point{Lat: 12.34, Lng: 56.78})
	require.Nil(t, err)

	id, err := result.LastInsertId()
	require.Nil(t, err)
	require.EqualValues(t, id, 1)

	var other Point
	err = pointSess.QueryRow(`SELECT location FROM points WHERE id = 1`).Scan(&other)
	require.Nil(t, err)

	require.EqualValues(t, other.Lat, 12.34)
	require.EqualValues(t, other.Lng, 56.78)
}
