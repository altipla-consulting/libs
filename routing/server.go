package routing

import (
	"context"
	"net/http"
	"text/template"

	"github.com/julienschmidt/httprouter"
	log "github.com/sirupsen/logrus"

	"libs.altipla.consulting/errors"
	"libs.altipla.consulting/langs"
	"libs.altipla.consulting/sentry"
)

// Handler should be implemented by the handler functions that we want to register.
type Handler func(w http.ResponseWriter, r *http.Request) error

// ServerOption is implement by any option that can be passed when constructing a new server.
type ServerOption func(server *Server)

// WithBetaAuth installs a rough authentication mechanism to avoid the final users
// to access beta sites.
func WithBetaAuth(username, password string) ServerOption {
	return func(server *Server) {
		server.username = username
		server.password = password
	}
}

// WithLogrus enables logging of the errors of the handlers.
func WithLogrus() ServerOption {
	return func(server *Server) {
		server.logging = true
	}
}

// WithSentry configures Sentry logging of issues in the handlers.
func WithSentry(dsn string) ServerOption {
	return func(server *Server) {
		if dsn != "" {
			server.sentryClient = sentry.NewClient(dsn)
		}
	}
}

// WithCustom404 uses a custom 404 template file.
func WithCustom404(handler Handler) ServerOption {
	return func(server *Server) {
		server.handler404 = handler
	}
}

// Server configures the routing table.
type Server struct {
	router *httprouter.Router

	// Options
	username, password string
	sentryClient       *sentry.Client
	logging            bool
	handler404         Handler
}

// NewServer configures a new router with the options.
func NewServer(opts ...ServerOption) *Server {
	s := &Server{
		router: httprouter.New(),
	}
	for _, opt := range opts {
		opt(s)
	}

	if s.handler404 == nil {
		s.handler404 = s.generic404Handler
	}
	s.router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.decorate(langs.ES, s.handler404)(w, r, nil)
	})

	return s
}

func (s *Server) generic404Handler(w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusNotFound)

	tmpl, err := template.New("error").Parse(errorTemplate)
	if err != nil {
		return errors.Wrapf(err, "cannot parse default 404 error template")
	}
	if err := tmpl.Execute(w, http.StatusNotFound); err != nil {
		return errors.Wrapf(err, "cannot execute default 404 error template")
	}

	return nil
}

// Router returns the raw underlying router to make advanced modifications.
// If you modify the NotFound handler remember to call routing.NotFoundHandler
// as fallback if you don't want to process the request.
func (s *Server) Router() *httprouter.Router {
	return s.router
}

// Get registers a new GET route in the router.
func (s *Server) Get(lang, path string, handler Handler) {
	s.router.GET(path, s.decorate(lang, handler))
}

// Post registers a new POST route in the router.
func (s *Server) Post(lang, path string, handler Handler) {
	s.router.POST(path, s.decorate(lang, handler))
}

// Put registers a new PUT route in the router.
func (s *Server) Put(lang, path string, handler Handler) {
	s.router.PUT(path, s.decorate(lang, handler))
}

// Delete registers a new DELETE route in the router.
func (s *Server) Delete(lang, path string, handler Handler) {
	s.router.DELETE(path, s.decorate(lang, handler))
}

// Group registers all the routes of the group in the router.
func (s *Server) Group(g Group) {
	for lang, url := range g.URL {
		h := func(w http.ResponseWriter, r *http.Request) error {
			r = r.WithContext(context.WithValue(r.Context(), groupKey, g))

			return g.Handler(w, r)
		}

		switch g.Method {
		case http.MethodGet:
			s.Get(lang, url, h)
		case http.MethodPost:
			s.Post(lang, url, h)
		case http.MethodDelete:
			s.Delete(lang, url, h)
		default:
			s.Get(lang, url, h)
		}
	}
}

func (s *Server) decorate(lang string, handler Handler) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		r = r.WithContext(context.WithValue(r.Context(), requestKey, r))
		r = r.WithContext(context.WithValue(r.Context(), paramsKey, ps))
		r = r.WithContext(context.WithValue(r.Context(), langKey, lang))

		if s.username != "" && s.password != "" {
			if _, err := r.Cookie("routing.beta"); err != nil && err != http.ErrNoCookie {
				log.WithField("error", err.Error()).Error("Cannot read cookie")
				s.emitError(w, r, http.StatusInternalServerError)
				return
			} else if err == http.ErrNoCookie {
				w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)

				username, password, ok := r.BasicAuth()
				if !ok {
					s.emitError(w, r, http.StatusUnauthorized)
					return
				}
				if username != s.username || password != s.password {
					s.emitError(w, r, http.StatusUnauthorized)
					return
				}

				http.SetCookie(w, &http.Cookie{
					Name:  "routing.beta",
					Value: "beta",
				})
			}
		}

		if err := handler(w, r); err != nil {
			if httperr, ok := err.(Error); ok {
				switch httperr.StatusCode {
				case http.StatusNotFound, http.StatusUnauthorized, http.StatusBadRequest:
					s.emitError(w, r, httperr.StatusCode)
					return
				}
			}

			if s.logging {
				log.WithFields(log.Fields{
					"error":   err.Error(),
					"details": errors.Details(err),
				}).Errorf("Handler failed")
			}

			if s.sentryClient != nil {
				s.sentryClient.ReportRequest(err, r)
			}

			s.emitError(w, r, http.StatusInternalServerError)
		}
	}
}

func (s *Server) emitError(w http.ResponseWriter, r *http.Request, status int) {
	if status == http.StatusNotFound {
		s.router.NotFound.ServeHTTP(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(status)

	tmpl, err := template.New("error").Parse(errorTemplate)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.WithField("error", err.Error()).Error("Cannot parse template")
	}
	if err := tmpl.Execute(w, status); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.WithField("error", err.Error()).Error("Cannot execute template")
	}
}
