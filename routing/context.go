package routing

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
)

type key int

const (
	requestKey key = 1
)

// Param returns a request URL parameter value.
func Param(r *http.Request, name string) string {
	vars := mux.Vars(r)
	return vars["name"]
}

// RequestFromContext returns the HTTP request from a context. The context must
// have been previously extracted from r.Context().
func RequestFromContext(ctx context.Context) *http.Request {
	return ctx.Value(requestKey).(*http.Request)
}
