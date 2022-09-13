package doris

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"cloud.google.com/go/profiler"
	"github.com/sethvargo/go-signalcontext"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"libs.altipla.consulting/env"
	"libs.altipla.consulting/errors"
	"libs.altipla.consulting/routing"
)

type Server struct {
	*routing.Server
	internal *routing.Server
	cnf      *config
	ctx      context.Context
	cancel   context.CancelFunc
}

func NewServer(opts ...Option) *Server {
	cnf := &config{
		http: []routing.ServerOption{
			routing.WithLogrus(),
			routing.WithSentry(os.Getenv("SENTRY_DSN")),
		},
	}
	for _, opt := range opts {
		opt(cnf)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Server{
		Server:   routing.NewServer(cnf.http...),
		internal: routing.NewServer(routing.WithLogrus()),
		cnf:      cnf,
		ctx:      ctx,
		cancel:   cancel,
	}
}

func (server *Server) Context() context.Context {
	return server.ctx
}

func (server *Server) port() string {
	if port := os.Getenv("PORT"); port != "" {
		return port
	}
	return "8080"
}

func healthHandler(w http.ResponseWriter, r *http.Request) error {
	fmt.Fprintf(w, "%s is ok\n", env.ServiceName())
	return nil
}

func (server *Server) Serve() {
	signalctx, done := signalcontext.OnInterrupt()
	defer done()

	if os.Getenv("SENTRY_DSN") != "" {
		log.WithField("dsn", os.Getenv("SENTRY_DSN")).Info("Sentry enabled")
	}

	if server.cnf.profiler {
		log.Info("Stackdriver Profiler enabled")

		cnf := profiler.Config{
			Service:        env.ServiceName(),
			ServiceVersion: env.Version(),
		}
		if err := profiler.Start(cnf); err != nil {
			log.Fatalf("failed to configure profiler: %s", err)
		}
	}

	server.Get("/health", healthHandler)
	server.internal.Get("/health", healthHandler)

	web := &http.Server{
		Addr:    ":" + server.port(),
		Handler: h2c.NewHandler(server, new(http2.Server)),
	}
	internal := &http.Server{
		Addr:    ":8000",
		Handler: server.internal,
	}

	go func() {
		if err := web.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("failed to serve: %s", err)
		}
	}()
	go func() {
		if err := internal.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("failed to serve: %s", err)
		}
	}()

	log.WithFields(log.Fields{
		"port":          server.port(),
		"internal-port": ":8000",
		"version":       env.Version(),
		"name":          env.ServiceName(),
	}).Info("Instance initialized successfully!")

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()

		<-signalctx.Done()

		log.Info("Shutting down")

		server.cancel()

		shutdownctx, done := context.WithTimeout(context.Background(), 25*time.Second)
		defer done()

		_ = internal.Shutdown(shutdownctx)
		_ = web.Shutdown(shutdownctx)
	}()
	wg.Wait()
}
