package routing

import (
	"net/http"
	"strings"
)

// Group registers multiple URLs for each lang for the same handler.
type Group struct {
	// Handler that will be called for all langs.
	Handler Handler

	// HTTP method to use. If empty or unknown it will use "GET".
	Method string

	// Map of language -> URL that should be registered.
	URL map[string]string
}

// ResolveURL returns the URL of the group linked to the language we pass. If no
// URL is found it will return an empty string. It replaces any parameter with
// the value passed right now in the request.
func (g Group) ResolveURL(r *http.Request, lang string) string {
	segments := strings.Split(g.URL[lang], "/")
	for i, segment := range segments {
		if strings.HasPrefix(segment, ":") {
			segments[i] = Param(r, segment[1:])
		}
	}

	return strings.Join(segments, "/")
}
