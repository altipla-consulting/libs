package doris

import (
	"net/http"

	"github.com/rs/cors"

	"libs.altipla.consulting/routing"
)

type RegisterFn func() (pattern string, handler http.Handler)

func Connect(r *routing.Router, fn RegisterFn) {
	pattern, handler := fn()
	r.PathPrefixHandler(pattern, routing.NewHandlerFromHTTP(handler))
}

// ConnectCORS returns a CORS configuration for the given domains with the
// optimal settings for a Connect API.
func ConnectCORS(origins []string) cors.Options {
	return cors.Options{
		AllowedOrigins: origins,
		AllowedMethods: []string{http.MethodPost, http.MethodOptions},
		AllowedHeaders: []string{"authorization", "content-type"},
		MaxAge:         300,
	}
}
