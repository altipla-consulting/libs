package pagination

import "github.com/altipla-consulting/errors"

var (
	ErrInvalidToken     = errors.New("pagination: invalid token")
	ErrChecksumMismatch = errors.New("pagination: checksum mismatch")
)
