package errors

import (
	"database/sql"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestGRPCError(t *testing.T) {
	err := status.Errorf(codes.NotFound, "foo not found")
	err = Trace(err)
	err = Wrapf(err, "bar")

	s, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, s.Code(), codes.NotFound)
	require.Equal(t, s.Message(), "foo not found")
}

func TestUnknownErrorAsGRPC(t *testing.T) {
	err := New("unrelated")
	err = Trace(err)
	err = Wrapf(err, "bar")

	s, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, s.Code(), codes.Unknown)
	require.Equal(t, s.Message(), "unrelated")
}

func TestWrappingNativeStackedErrors(t *testing.T) {
	err := Trace(fmt.Errorf("cannot query: %w", sql.ErrNoRows))

	require.True(t, errors.Is(err, sql.ErrNoRows))
	require.True(t, Is(err, sql.ErrNoRows))
}
