//go:build go1.13
// +build go1.13

package errors

import (
	"errors" // revive:disable-line:imports-blacklist
)

// As finds the first error in err's chain that matches target, and if so, sets
// target to that error value and returns true.
func As(err error, target interface{}) bool {
	return errors.As(err, target)
}
