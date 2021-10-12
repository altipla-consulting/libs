package cloudrun

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/profiler"
	"github.com/sethvargo/go-signalcontext"
	log "github.com/sirupsen/logrus"

	"libs.altipla.consulting/env"
	"libs.altipla.consulting/errors"
	"libs.altipla.consulting/routing"
)

type WebServer struct {
	*routing.Server
	cnf *config
}

func Web(opts ...Option) *WebServer {
	cnf := &config{
		http: []routing.ServerOption{
			routing.WithLogrus(),
			routing.WithSentry(os.Getenv("SENTRY_DSN")),
		},
	}
	for _, opt := range opts {
		opt(cnf)
	}

	return &WebServer{
		Server: routing.NewServer(cnf.http...),
		cnf:    cnf,
	}
}

func (server *WebServer) port() string {
	if port := os.Getenv("PORT"); port != "" {
		return port
	}
	return "8080"
}

func (server *WebServer) Serve() {
	ctx, done := signalcontext.OnInterrupt()
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
			log.Fatal(err)
		}
	}

	server.Get("/health", func(w http.ResponseWriter, r *http.Request) error {
		fmt.Fprintf(w, "%s is ok\n", env.ServiceName())
		return nil
	})

	lis, err := net.Listen("tcp", ":"+server.port())
	if err != nil {
		log.Fatal(err)
	}

	web := &http.Server{
		Handler: server,
	}

	errch := make(chan error, 1)
	go func() {
		<-ctx.Done()

		shutdownctx, done := context.WithTimeout(context.Background(), 7*time.Second)
		defer done()

		log.Info("Shutting down")
		if err := web.Shutdown(shutdownctx); err != nil {
			select {
			case errch <- err:
			default:
			}
		}
	}()

	log.WithFields(log.Fields{
		"port":    server.port(),
		"version": env.Version(),
		"name":    env.ServiceName(),
	}).Info("Instance initialized successfully!")

	// Run the server. This will block until the context is closed.
	if err := web.Serve(lis); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("failed to serve: %s", err)
	}

	// Return any errors that happened during shutdown.
	select {
	case err := <-errch:
		log.Fatalf("failed to shutdown: %s", err)
	default:
	}
}
