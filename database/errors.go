package database

import (
	"errors"
	"strings"
)

var (
	// ErrNoSuchEntity is returned from a Get operation when there is not a model
	// that matches the query
	ErrNoSuchEntity = errors.New("database: no such entity")

	// ErrDone is returned from the Next() method of an iterator when all results
	// have been read.
	ErrDone = errors.New("query has no more results")

	// ErrConcurrentTransaction is returned when trying to update a model that has been
	// updated in the background by other process. This errors will prevent you from
	// potentially overwriting those changes.
	ErrConcurrentTransaction = errors.New("database: concurrent transaction")
)

// MultiError stores a list of error when retrieving multiple models and only
// some of them may fail.
type MultiError []error

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

// HasError returns if the multi error really contains any error or all the rows
// have been successfully retrieved. You don't have to check this method most of
// the time because GetMulti, GetAll, etc. will return nil if there is no errors
// instead of an empty MultiErrorto avoid hard to debug bugs.
func (merr MultiError) HasError() bool {
	for _, err := range merr {
		if err != nil {
			return true
		}
	}

	return false
}
