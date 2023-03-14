package rdb

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/altipla-consulting/errors"
)

var (
	ErrNoSuchEntity          = errors.New("rdb: no such entity")
	ErrConcurrentTransaction = errors.New("rdb: concurrent transaction")
	ErrDatabaseDoesNotExists = errors.New("rdb: database does not exists")
)

type noSuchEntityError struct {
	reason string
}

func newNoSuchEntityError(format string, args ...any) noSuchEntityError {
	return noSuchEntityError{fmt.Sprintf(format, args...)}
}

func (err noSuchEntityError) Error() string {
	return fmt.Sprintf("%s: %v", ErrNoSuchEntity, err.reason)
}

func (err noSuchEntityError) Unwrap() error {
	return ErrNoSuchEntity
}

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
// instead of an empty MultiError to avoid hard to debug bugs.
func (merr MultiError) HasError() bool {
	for _, err := range merr {
		if err != nil {
			return true
		}
	}

	return false
}

type UnexpectedStatusError struct {
	RequestMethod string
	RequestURL    *url.URL
	Status        string
	StatusCode    int
	Advanced      *AdvancedError
}

func (err UnexpectedStatusError) Error() string {
	if err.Advanced != nil {
		return "unexpected error: " + err.Advanced.Error()
	}
	return fmt.Sprintf("unexpected status code %s calling %s %s", err.Status, err.RequestMethod, err.RequestURL.String())
}

type AdvancedError struct {
	URL     string `json:"Url"`
	Type    string
	Message string
	Stack   string `json:"Error"`
}

func (err *AdvancedError) Error() string {
	return fmt.Sprintf("%s: %s", err.Type, err.Message)
}

func NewUnexpectedStatusError(r *http.Request, resp *http.Response) UnexpectedStatusError {
	unexpected := UnexpectedStatusError{
		RequestMethod: r.Method,
		RequestURL:    r.URL,
		Status:        resp.Status,
		StatusCode:    resp.StatusCode,
	}
	if resp.StatusCode == http.StatusInternalServerError || resp.StatusCode == http.StatusBadRequest {
		// Take care, checks if there is NO error
		advanced := new(AdvancedError)
		if err := json.NewDecoder(resp.Body).Decode(&advanced); err == nil {
			unexpected.Advanced = advanced
		}
	}

	return unexpected
}
