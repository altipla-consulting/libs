package pagination

import "libs.altipla.consulting/errors"

var (
	ErrInvalidToken     = errors.New("pagination: invalid token")
	ErrChecksumMismatch = errors.New("pagination: checksum mismatch")
)
