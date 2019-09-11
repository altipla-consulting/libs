package loaders

import (
	"sync"

	"libs.altipla.consulting/errors"
)

// Lazy runs the function only once and returns the error if there is one. If the
// function has already been executed it will never return an error.
//
// It is safe for concurrent use as it uses the once provided to it.
func Lazy(once *sync.Once, f func() error) error {
	var err error
	once.Do(func() {
		err = errors.Trace(f())
	})
	return errors.Trace(err)
}
