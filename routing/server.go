package routing

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/altipla-consulting/env"
	"github.com/altipla-consulting/errors"
	"github.com/altipla-consulting/sentry"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

// Handler should be implemented by the handler functions that we want to register.
type Handler func(w http.ResponseWriter, r *http.Request) error

// NewHandlerFromHTTP creates a new handler from a standard http.Handler.
func NewHandlerFromHTTP(handler http.Handler) Handler {
	return func(w http.ResponseWriter, r *http.Request) error {
		handler.ServeHTTP(w, r)
		return nil
	}
}

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

// WithCustomTimeout changes the default timeout of 29 seconds to a custom one.
// It's only recommended for environments where the limit is longer than 30-60 seconds. For example
// background queues might have 10 minutes of activity in some setups.
func WithCustomTimeout(timeout time.Duration) ServerOption {
	return func(server *Server) {
		server.timeout = timeout
	}
}

// Server configures the routing table.
type Server struct {
	*Router

	// Options
	username, password string
	sentryClient       *sentry.Client
	logging            bool
	handler404         Handler
	timeout            time.Duration
}

// NewServer configures a new router with the options.
func NewServer(opts ...ServerOption) *Server {
	s := &Server{
		timeout: 29 * time.Second,
	}
	for _, opt := range opts {
		opt(s)
	}
	if s.handler404 == nil {
		s.handler404 = generic404Handler
	}

	s.Router = &Router{
		s: s,
		r: mux.NewRouter().StrictSlash(true),
	}
	s.r.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.decorate(nil, s.handler404)(w, r)
	})

	return s
}

func (s *Server) call404(w http.ResponseWriter, r *http.Request) {
	s.decorate(nil, s.handler404)(w, r)
}

func generic404Handler(w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusNotFound)

	tmpl, err := template.New("error").Parse(errorTemplate)
	if err != nil {
		return fmt.Errorf("cannot parse default 404 error template: %v", err)
	}
	if err := tmpl.Execute(w, http.StatusNotFound); err != nil {
		return fmt.Errorf("cannot execute default 404 error template: %v", err)
	}

	return nil
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.Router.r.ServeHTTP(w, r)
}

func (s *Server) decorate(middlewares []Middleware, handler Handler) http.HandlerFunc {
	for _, middleware := range middlewares {
		handler = middleware(handler)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		defer s.sentryClient.ReportPanicsRequest(r)

		ctx := r.Context()
		ctx = context.WithValue(ctx, requestKey, r)
		ctx, cancel := context.WithTimeout(ctx, s.timeout)
		defer cancel()

		r = r.WithContext(ctx)

		if s.username != "" && s.password != "" {
			if _, err := r.Cookie("routing.beta"); err != nil && err != http.ErrNoCookie {
				slog.Error("Cannot read beta auth cookie", slog.String("error", err.Error()))
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
					slog.Error("Handler failed",
						slog.String("status", http.StatusText(herr.StatusCode)),
						slog.String("reason", herr.Message),
						slog.String("error", err.Error()))
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
				slog.Error("Handler failed", slog.String("error", err.Error()))
			}
			s.sentryClient.ReportRequest(r, err)

			// Responde según el tipo de error por timeout u otro con un código HTTP adecuado.
			if ctx.Err() == context.DeadlineExceeded {
				s.emitError(w, r, http.StatusGatewayTimeout)
				return
			}
			if env.IsLocal() {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintln(w, errors.Stack(err))
				return
			}
			s.emitError(w, r, http.StatusInternalServerError)
		}
	}
}

func (s *Server) emitError(w http.ResponseWriter, r *http.Request, status int) {
	if status == http.StatusNotFound {
		s.call404(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(status)

	tmpl, err := template.New("error").Parse(errorTemplate)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		slog.Error("Cannot parse error template", slog.String("error", err.Error()))
	}
	if err := tmpl.Execute(w, status); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		slog.Error("Cannot execute error template", slog.String("error", err.Error()))
	}
}

type Router struct {
	s           *Server
	r           *mux.Router
	middlewares []Middleware
}

type Middleware func(handler Handler) Handler

// Get registers a new GET route.
func (router *Router) Get(path string, handler Handler) {
	fn := router.s.decorate(router.middlewares, handler)
	if prefix := hasWildcard(path); prefix != "" {
		router.r.PathPrefix(prefix).HandlerFunc(fn).Methods(http.MethodGet)
		router.r.PathPrefix(prefix).HandlerFunc(fn).Methods(http.MethodHead)
		return
	}
	router.r.HandleFunc(router.migratePath(path), fn).Methods(http.MethodGet)
	router.r.HandleFunc(router.migratePath(path), fn).Methods(http.MethodHead)
}

// Post registers a new POST route.
func (router *Router) Post(path string, handler Handler) {
	fn := router.s.decorate(router.middlewares, handler)
	if prefix := hasWildcard(path); prefix != "" {
		router.r.PathPrefix(prefix).HandlerFunc(fn).Methods(http.MethodPost)
		return
	}
	router.r.HandleFunc(router.migratePath(path), fn).Methods(http.MethodPost)
}

// Put registers a new PUT route.
func (router *Router) Put(path string, handler Handler) {
	fn := router.s.decorate(router.middlewares, handler)
	if prefix := hasWildcard(path); prefix != "" {
		router.r.PathPrefix(prefix).HandlerFunc(fn).Methods(http.MethodPut)
		return
	}
	router.r.HandleFunc(router.migratePath(path), fn).Methods(http.MethodPut)
}

// Delete registers a new DELETE route.
func (router *Router) Delete(path string, handler Handler) {
	fn := router.s.decorate(router.middlewares, handler)
	if prefix := hasWildcard(path); prefix != "" {
		router.r.PathPrefix(prefix).HandlerFunc(fn).Methods(http.MethodDelete)
		return
	}
	router.r.HandleFunc(router.migratePath(path), fn).Methods(http.MethodDelete)
}

// Options registers a new OPTIONS route.
func (router *Router) Options(path string, handler Handler) {
	fn := router.s.decorate(router.middlewares, handler)
	if prefix := hasWildcard(path); prefix != "" {
		router.r.PathPrefix(prefix).HandlerFunc(fn).Methods(http.MethodOptions)
		return
	}
	router.r.HandleFunc(router.migratePath(path), fn).Methods(http.MethodOptions)
}

// Head registers a new HEAD route.
func (router *Router) Head(path string, handler Handler) {
	fn := router.s.decorate(router.middlewares, handler)
	if prefix := hasWildcard(path); prefix != "" {
		router.r.PathPrefix(prefix).HandlerFunc(fn).Methods(http.MethodHead)
		return
	}
	router.r.HandleFunc(router.migratePath(path), fn).Methods(http.MethodHead)
}

func (router *Router) PathPrefixHandler(pathPrefix string, handler Handler) {
	fn := router.s.decorate(router.middlewares, handler)
	router.r.PathPrefix(pathPrefix).HandlerFunc(fn)
}

func hasWildcard(path string) string {
	if strings.Contains(path, "*filepath") {
		return strings.TrimSuffix(path, "*filepath")
	}
	return ""
}

func (router *Router) migratePath(path string) string {
	parts := strings.Split(path, "/")
	for i, part := range parts {
		if strings.HasPrefix(part, ":") {
			parts[i] = "{" + part[1:] + "}"
		}
	}
	return strings.Join(parts, "/")
}

// ServeFiles register a raw net/http handler with no error checking that sends files.
func (router *Router) ServeFiles(path string, root http.FileSystem) {
	router.r.PathPrefix(path).Handler(http.StripPrefix(path, http.FileServer(root)))
}

func (router *Router) ProxyLocalAssets(destAddress string) {
	if !env.IsLocal() {
		return
	}

	u, err := url.Parse(destAddress)
	if err != nil {
		slog.Error("Cannot parse proxy address",
			slog.String("address", destAddress),
			slog.String("error", err.Error()))
		os.Exit(1)
	}
	proxy := httputil.NewSingleHostReverseProxy(u)
	router.r.PathPrefix("/static/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		proxy.ServeHTTP(w, r)
	})
}

type RouterOption func(router *Router)

// WithMiddleware adds a middleware to the router.
func WithMiddleware(middleware Middleware) RouterOption {
	return func(router *Router) {
		router.middlewares = append(router.middlewares, middleware)
	}
}

// WithCORS configures a CORS middleware to authorize cross-origin requests.
func WithCORS(opts cors.Options) RouterOption {
	return func(router *Router) {
		router.r.Use(cors.New(opts).Handler)
	}
}

// Domain matches a hostname.
func (router *Router) Domain(host string, opts ...RouterOption) *Router {
	sub := &Router{
		s:           router.s,
		r:           router.r.Host(host).Subrouter(),
		middlewares: router.middlewares,
	}
	for _, opt := range opts {
		opt(sub)
	}
	return sub
}

// Header matches a header with a specific value. If the value is empty it will match
// any value as long as the header is present in the request.
func (router *Router) Header(header, value string, opts ...RouterOption) *Router {
	sub := &Router{
		s:           router.s,
		r:           router.r.Headers(header, value).Subrouter(),
		middlewares: router.middlewares,
	}
	for _, opt := range opts {
		opt(sub)
	}
	return sub
}

// PathPrefix matches a path prefix.
func (router *Router) PathPrefix(path string, opts ...RouterOption) *Router {
	sub := &Router{
		s:           router.s,
		r:           router.r.PathPrefix(path).Subrouter(),
		middlewares: router.middlewares,
	}
	for _, opt := range opts {
		opt(sub)
	}
	return sub
}
