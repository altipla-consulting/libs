package grpctest

import (
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// RequireError assets that the GRPC error is the expected one.
func RequireError(t *testing.T, err error, code codes.Code, message string) {
	statusCode, _ := status.FromError(err)
	require.Equal(t, statusCode.Code(), code)
	require.Equal(t, statusCode.Message(), message)
}
