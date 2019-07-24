package routing

import (
	"fmt"
	"net/http"
)

// Error stores info about a HTTP error returned from a route.
type Error struct {
	StatusCode int
	Message    string
}

// Error implements the error interface with the message.
func (err Error) Error() string {
	return fmt.Sprintf("routing error %d: %s", err.StatusCode, err.Message)
}

// NotFound returns a 404 HTTP error and formats its message.
func NotFound(s string, args ...interface{}) error {
	return Error{
		StatusCode: http.StatusNotFound,
		Message:    fmt.Sprintf(s, args...),
	}
}

// Unauthorized returns a 401 HTTP error and formats its message.
func Unauthorized(s string, args ...interface{}) error {
	return Error{
		StatusCode: http.StatusUnauthorized,
		Message:    fmt.Sprintf(s, args...),
	}
}

// BadRequest returns a 400 HTTP error and formats its message.
func BadRequest(s string, args ...interface{}) error {
	return Error{
		StatusCode: http.StatusBadRequest,
		Message:    fmt.Sprintf(s, args...),
	}
}

// Internal returns a 500 HTTP error and formats its message.
func Internal(s string, args ...interface{}) error {
	return Error{
		StatusCode: http.StatusInternalServerError,
		Message:    fmt.Sprintf(s, args...),
	}
}
