package hosting

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/profiler"
	"github.com/altipla-consulting/env"
	"github.com/altipla-consulting/errors"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/rs/cors"
	"github.com/sethvargo/go-signalcontext"
	log "github.com/sirupsen/logrus"
	"github.com/soheilhy/cmux"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"

	"libs.altipla.consulting/routing"
)

type GRPCServer struct {
	*grpc.Server
	http     *routing.Server
	gateway  *runtime.ServeMux
	cnf      *config
	platform Platform
}

func GRPC(platform Platform, opts ...Option) *GRPCServer {
	cnf := &config{
		http: []routing.ServerOption{
			routing.WithSentry(os.Getenv("SENTRY_DSN")),
		},
		grpc: []grpc.ServerOption{},
		unaryInterceptors: []grpc.UnaryServerInterceptor{
			grpcUnaryErrorLogger(),
			grpcTrimStrings(),
		},
	}
	for _, opt := range opts {
		opt(cnf)
	}

	cnf.grpc = append(cnf.grpc, grpc.ChainUnaryInterceptor(cnf.unaryInterceptors...))

	server := &GRPCServer{
		Server:   grpc.NewServer(cnf.grpc...),
		http:     routing.NewServer(cnf.http...),
		cnf:      cnf,
		platform: platform,
	}
	fn := func(ctx context.Context, mux *runtime.ServeMux, marshaler runtime.Marshaler, w http.ResponseWriter, r *http.Request, httpStatus int) {
		if httpStatus == http.StatusNotFound {
			server.http.ServeHTTP(w, r)
		} else {
			runtime.DefaultRoutingErrorHandler(ctx, mux, marshaler, w, r, httpStatus)
		}
	}
	server.gateway = runtime.NewServeMux(
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{MarshalOptions: protojson.MarshalOptions{EmitUnpopulated: true}}),
		runtime.WithRoutingErrorHandler(fn),
	)

	return server
}

func (server *GRPCServer) port() string {
	if port := os.Getenv("PORT"); port != "" {
		return port
	}
	return "8080"
}

func (server *GRPCServer) Gateway(fn func(context.Context, *runtime.ServeMux, string, []grpc.DialOption) error) {
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithDefaultCallOptions(grpc.WaitForReady(true)),
	}
	if err := fn(context.Background(), server.gateway, "localhost:"+server.port(), opts); err != nil {
		log.Fatal(err)
	}
}

func (server *GRPCServer) HTTPGateway(fn func(r *routing.Server)) {
	fn(server.http)
}

func (server *GRPCServer) Serve() {
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

	server.http.Get("/health", func(w http.ResponseWriter, r *http.Request) error {
		fmt.Fprintf(w, "%s is ok\n", env.ServiceName())
		return nil
	})

	lis, err := net.Listen("tcp", ":"+server.port())
	if err != nil {
		log.Fatal(err)
	}
	mux := cmux.New(lis)
	lisGRPC := mux.MatchWithWriters(cmux.HTTP2MatchHeaderFieldSendSettings("content-type", "application/grpc"))
	lisHTTP := mux.Match(cmux.Any())

	c := cors.New(cors.Options{
		AllowedOrigins: server.cnf.cors,
		AllowedMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
		AllowedHeaders: []string{"authorization", "content-type"},
		ExposedHeaders: []string{"grpc-message", "grpc-status"},
		MaxAge:         300,
	})
	web := &http.Server{
		Handler: c.Handler(server.gateway),
	}

	errch := make(chan error, 1)
	go func() {
		<-ctx.Done()

		log.Info("Shutting down")

		shutdownctx, done := context.WithTimeout(context.Background(), 25*time.Second)
		defer done()

		if err := server.platform.Shutdown(shutdownctx); err != nil {
			select {
			case errch <- err:
			default:
			}
		}
		if err := web.Shutdown(shutdownctx); err != nil {
			select {
			case errch <- err:
			default:
			}
		}
		server.Server.GracefulStop()
		if err := lis.Close(); err != nil {
			select {
			case errch <- err:
			default:
			}
		}
	}()

	if err := server.platform.Init(); err != nil {
		log.Fatalf("cannot initialize platform: %s", err)
	}

	log.WithFields(log.Fields{
		"port":    server.port(),
		"version": env.Version(),
		"name":    env.ServiceName(),
	}).Info("Instance initialized successfully!")

	go server.Server.Serve(lisGRPC)
	go web.Serve(lisHTTP)

	// Run the server. This will block until the context is closed.
	if err := mux.Serve(); err != nil {
		var neterr *net.OpError
		if errors.As(err, &neterr) {
			if neterr.Err.Error() != "use of closed network connection" {
				log.Fatalf("failed to serve: %s", err)
			}
		} else {
			log.Fatalf("failed to serve: %s", err)
		}
	}

	// Return any errors that happened during shutdown.
	select {
	case err := <-errch:
		log.Fatalf("failed to shutdown: %s", err)
	default:
	}
}
