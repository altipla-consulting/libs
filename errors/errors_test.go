package errors

import (
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
