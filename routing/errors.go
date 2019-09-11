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

// NotFound returns a 404 HTTP error.
func NotFound(s string) error {
	return Error{
		StatusCode: http.StatusNotFound,
		Message:    s,
	}
}

// NotFoundf returns a 404 HTTP error and formats its message.
func NotFoundf(s string, args ...interface{}) error {
	return Error{
		StatusCode: http.StatusNotFound,
		Message:    fmt.Sprintf(s, args...),
	}
}

// Unauthorized returns a 401 HTTP error.
func Unauthorized(s string) error {
	return Error{
		StatusCode: http.StatusUnauthorized,
		Message:    s,
	}
}

// Unauthorizedf returns a 401 HTTP error and formats its message.
func Unauthorizedf(s string, args ...interface{}) error {
	return Error{
		StatusCode: http.StatusUnauthorized,
		Message:    fmt.Sprintf(s, args...),
	}
}

// BadRequest returns a 400 HTTP error.
func BadRequest(s string) error {
	return Error{
		StatusCode: http.StatusBadRequest,
		Message:    s,
	}
}

// BadRequestf returns a 400 HTTP error and formats its message.
func BadRequestf(s string, args ...interface{}) error {
	return Error{
		StatusCode: http.StatusBadRequest,
		Message:    fmt.Sprintf(s, args...),
	}
}

// Internal returns a 500 HTTP error.
func Internal(s string) error {
	return Error{
		StatusCode: http.StatusInternalServerError,
		Message:    s,
	}
}

// Internalf returns a 500 HTTP error and formats its message.
func Internalf(s string, args ...interface{}) error {
	return Error{
		StatusCode: http.StatusInternalServerError,
		Message:    fmt.Sprintf(s, args...),
	}
}
