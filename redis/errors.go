package redis

import (
	"strings"

	"github.com/altipla-consulting/errors"
)

var (
	// ErrNoSuchEntity is returned from a Get operation when there is not a model
	// that matches the query
	ErrNoSuchEntity = errors.New("no such entity")

	ErrDone = errors.New("done")
)

// MultiError is returned from batch operations with the error of each operation.
// If a batch operation does not fails this will be nil too.
type MultiError []error

// Error returns the composed error message with all the individual ones.
func (merr MultiError) Error() string {
	var msg []string
	for _, err := range merr {
		if err == nil {
			msg = append(msg, "<nil>")
		} else {
			msg = append(msg, err.Error())
		}
	}

	return strings.Join(msg, "; ")
}

// HasError checks if there is really an error on the list or all of them are empty.
func (merr MultiError) HasError() bool {
	for _, err := range merr {
		if err != nil {
			return true
		}
	}

	return false
}
