package routing

import (
	"context"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type key int

const (
	requestKey key = 1
	paramsKey  key = 2
)

// Param returns a request URL parameter value.
func Param(r *http.Request, name string) string {
	return r.Context().Value(paramsKey).(httprouter.Params).ByName(name)
}

// RequestFromContext returns the HTTP request from a context. The context must
// have been previously extracted from r.Context().
func RequestFromContext(ctx context.Context) *http.Request {
	return ctx.Value(requestKey).(*http.Request)
}
