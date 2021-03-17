package crypt

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	pbtimestamp "google.golang.org/protobuf/types/known/timestamppb"

	"libs.altipla.consulting/datetime"
)

func TestTokenDoesNotContainsBase64Padding(t *testing.T) {
	s := NewSigner("01234567890123456789012345678912", "1234567890123456")
	msg := datetime.SerializeTimestamp(time.Date(2006, time.January, 2, 3, 4, 5, 0, time.UTC))
	token, err := s.SignMessage(msg)
	require.NoError(t, err)

	require.False(t, strings.Contains(token, "="))

	decoded := new(pbtimestamp.Timestamp)
	require.NoError(t, s.ReadMessage(token, decoded))

	require.WithinDuration(t, datetime.ParseTimestamp(decoded), time.Date(2006, time.January, 2, 3, 4, 5, 0, time.UTC), 1*time.Second)
}
