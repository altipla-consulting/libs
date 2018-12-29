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
	langKey    key = 3
	groupKey   key = 4
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

// GroupFromContext returns the context of the request. If the route was direct
// and used no group it returns an empty group with no URLs.
func GroupFromContext(ctx context.Context) Group {
	return ctx.Value(groupKey).(Group)
}

// LangFromContext returns the lang of the route linked to the request in the routing
// table. The context must have been previously extracted from r.Context().
func LangFromContext(ctx context.Context) string {
	return ctx.Value(langKey).(string)
}

// Lang returns the language linked to the request in the routing table.
func Lang(r *http.Request) string {
	return LangFromContext(r.Context())
}
