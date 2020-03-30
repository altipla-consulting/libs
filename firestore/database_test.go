package firestore

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func initDatabase(t *testing.T) *Database {
	db, err := Open("local")
	require.NoError(t, err)

	return db
}
