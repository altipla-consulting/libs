package routing

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"text/template"
	"time"

	"github.com/julienschmidt/httprouter"
	log "github.com/sirupsen/logrus"

	"libs.altipla.consulting/env"
	"libs.altipla.consulting/errors"
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
		server.sentryClient = sentry.NewClient(dsn)
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
	domains       map[string]*Domain
	defaultDomain *Domain

	// Options
	username, password string
	sentryClient       *sentry.Client
	logging            bool
	handler404         Handler
}

// NewServer configures a new router with the options.
func NewServer(opts ...ServerOption) *Server {
	s := &Server{
		domains: make(map[string]*Domain),
	}
	for _, opt := range opts {
		opt(s)
	}

	s.defaultDomain = &Domain{
		s:      s,
		router: httprouter.New(),
	}

	if s.handler404 == nil {
		s.handler404 = generic404Handler
	}
	s.defaultDomain.router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.decorate("NOTFOUND", "", s.handler404)(w, r, nil)
	})

	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if domain := s.domains[r.Host]; domain != nil {
		domain.router.ServeHTTP(w, r)
	} else {
		s.defaultDomain.router.ServeHTTP(w, r)
	}
}

func generic404Handler(w http.ResponseWriter, r *http.Request) error {
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

func (s *Server) Domain(host string) *Domain {
	if s.domains[host] == nil {
		s.domains[host] = &Domain{
			s:      s,
			router: httprouter.New(),
		}
		s.domains[host].router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			s.decorate("NOTFOUND", "", s.handler404)(w, r, nil)
		})
	}

	return s.domains[host]
}

type Domain struct {
	s      *Server
	router *httprouter.Router
}

// Get registers a new GET route in the domain.
func (domain *Domain) Get(path string, handler Handler) {
	domain.router.GET(path, domain.s.decorate(http.MethodGet, path, handler))
}

// Post registers a new POST route in the domain.
func (domain *Domain) Post(path string, handler Handler) {
	domain.router.POST(path, domain.s.decorate(http.MethodPost, path, handler))
}

// Put registers a new PUT route in the domain.
func (domain *Domain) Put(path string, handler Handler) {
	domain.router.PUT(path, domain.s.decorate(http.MethodPut, path, handler))
}

// Delete registers a new DELETE route in the domain.
func (domain *Domain) Delete(path string, handler Handler) {
	domain.router.DELETE(path, domain.s.decorate(http.MethodDelete, path, handler))
}

// Options registers a new OPTIONS route in the domain.
func (domain *Domain) Options(path string, handler Handler) {
	domain.router.OPTIONS(path, domain.s.decorate(http.MethodOptions, path, handler))
}

// ServeFiles register a raw net/http handler with no error checking that sends files.
func (domain *Domain) ServeFiles(path string, root http.FileSystem) {
	domain.router.ServeFiles(path, root)
}

func (domain *Domain) ProxyLocalAssets(destAddress string) {
	if !env.IsLocal() {
		return
	}

	u, err := url.Parse(destAddress)
	if err != nil {
		log.Fatal(err)
	}
	proxy := httputil.NewSingleHostReverseProxy(u)
	domain.Get("/static/*file", func(w http.ResponseWriter, r *http.Request) error {
		proxy.ServeHTTP(w, r)
		return nil
	})
}

// Get registers a new GET route in the router.
func (s *Server) Get(path string, handler Handler) {
	s.defaultDomain.Get(path, handler)
}

// Post registers a new POST route in the router.
func (s *Server) Post(path string, handler Handler) {
	s.defaultDomain.Post(path, handler)
}

// Put registers a new PUT route in the router.
func (s *Server) Put(path string, handler Handler) {
	s.defaultDomain.Put(path, handler)
}

// Delete registers a new DELETE route in the router.
func (s *Server) Delete(path string, handler Handler) {
	s.defaultDomain.Delete(path, handler)
}

// Options registers a new OPTIONS route in the router.
func (s *Server) Options(path string, handler Handler) {
	s.defaultDomain.Options(path, handler)
}

// ServeFiles register a raw net/http handler with no error checking that sends files.
func (s *Server) ServeFiles(path string, root http.FileSystem) {
	s.defaultDomain.ServeFiles(path, root)
}

func (s *Server) ProxyLocalAssets(destAddress string) {
	if !env.IsLocal() {
		return
	}

	u, err := url.Parse(destAddress)
	if err != nil {
		log.Fatal(err)
	}
	proxy := httputil.NewSingleHostReverseProxy(u)
	s.Get("/static/*file", func(w http.ResponseWriter, r *http.Request) error {
		proxy.ServeHTTP(w, r)
		return nil
	})
}

func (s *Server) decorate(method, path string, handler Handler) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		if s.sentryClient != nil {
			defer s.sentryClient.ReportPanicsRequest(r)
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, requestKey, r)
		ctx = context.WithValue(ctx, paramsKey, ps)
		ctx, cancel := context.WithTimeout(ctx, 29*time.Second)
		defer cancel()
		r = r.WithContext(ctx)

		if s.username != "" && s.password != "" {
			if _, err := r.Cookie("routing.beta"); err != nil && err != http.ErrNoCookie {
				log.WithField("error", err.Error()).Error("Cannot read beta auth cookie")
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
			var herr httpError
			if errors.As(err, &herr) {
				switch herr.StatusCode {
				case http.StatusNotFound, http.StatusUnauthorized, http.StatusBadRequest:
					log.WithFields(log.Fields{
						"status": http.StatusText(herr.StatusCode),
						"reason": herr.Message,
					}).Error("Handler failed")

					s.emitError(w, r, herr.StatusCode)
					return
				}
			}

			// Si el contexto se cancela simplemente mandamos el error al cliente ignorando
			// la respuesta desde el handler.
			if ctx.Err() == context.Canceled {
				s.emitError(w, r, http.StatusRequestTimeout)
				return
			}

			// Mandamos logging del error a la consola y/o Sentry.
			if s.logging {
				log.WithFields(errors.LogFields(err)).Error("Handler failed")
			}
			if s.sentryClient != nil {
				s.sentryClient.ReportRequest(r, err)
			}

			// Responde según el tipo de error por timeout u otro con un código HTTP adecuado.
			if ctx.Err() == context.DeadlineExceeded {
				s.emitError(w, r, http.StatusGatewayTimeout)
				return
			}
			if env.IsLocal() {
				fmt.Fprintln(w, errors.Stack(err))
				return
			}
			s.emitError(w, r, http.StatusInternalServerError)
		}
	}
}

func (s *Server) emitError(w http.ResponseWriter, r *http.Request, status int) {
	if status == http.StatusNotFound {
		s.defaultDomain.router.NotFound.ServeHTTP(w, r)
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
