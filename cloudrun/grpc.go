package cloudrun

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"cloud.google.com/go/profiler"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/rs/cors"
	"github.com/sethvargo/go-signalcontext"
	log "github.com/sirupsen/logrus"
	"github.com/soheilhy/cmux"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	"libs.altipla.consulting/env"
	"libs.altipla.consulting/errors"
	"libs.altipla.consulting/routing"
)

type GRPCServer struct {
	*grpc.Server
	http    *routing.Server
	gateway *runtime.ServeMux
	cnf     *config
}

func GRPC(opts ...Option) *GRPCServer {
	cnf := &config{
		http: []routing.ServerOption{
			routing.WithLogrus(),
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
		Server: grpc.NewServer(cnf.grpc...),
		http:   routing.NewServer(cnf.http...),
		cnf:    cnf,
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

	if cnf.queues != nil {
		cnf.queues(server.http)
	}
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

		shutdownctx, done := context.WithTimeout(context.Background(), 7*time.Second)
		defer done()

		log.Info("Shutting down")
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

func grpcTrimStrings() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		trimMessage(req.(proto.Message).ProtoReflect())
		return handler(ctx, req)
	}
}

func trimMessage(m protoreflect.Message) protoreflect.Message {
	m.Range(func(fd protoreflect.FieldDescriptor, value protoreflect.Value) bool {
		switch {
		case fd.Kind() == protoreflect.StringKind && fd.IsList():
			for list, i := value.List(), 0; i < list.Len(); i++ {
				list.Set(i, trimString(list.Get(i)))
			}

		case fd.Kind() == protoreflect.StringKind:
			m.Set(fd, trimString(value))

		// We need to trim the keys too, so we build a new map and then replace
		// the existing one once we ranged all the keys.
		case fd.Kind() == protoreflect.MessageKind && fd.IsMap() && fd.MapKey().Kind() == protoreflect.StringKind:
			trimmed := map[string]protoreflect.Value{}
			value.Map().Range(func(mk protoreflect.MapKey, v protoreflect.Value) bool {
				switch fd.MapValue().Kind() {
				case protoreflect.StringKind:
					v = trimString(v)
				case protoreflect.MessageKind:
					trimMessage(v.Message())
				}

				trimmed[trimString(mk.Value()).String()] = v
				value.Map().Clear(mk)
				return true
			})
			for k, sub := range trimmed {
				value.Map().Set(protoreflect.ValueOfString(k).MapKey(), sub)
			}

		case fd.Kind() == protoreflect.MessageKind && fd.IsMap():
			// Map has numeric keys, only process the values.
			value.Map().Range(func(mk protoreflect.MapKey, v protoreflect.Value) bool {
				switch fd.MapValue().Kind() {
				case protoreflect.StringKind:
					v.Map().Set(mk, trimString(v))
				case protoreflect.MessageKind:
					trimMessage(v.Message())
				}

				return true
			})

		case fd.Kind() == protoreflect.MessageKind && fd.IsList():
			for list, i := value.List(), 0; i < list.Len(); i++ {
				list.Set(i, protoreflect.ValueOfMessage(trimMessage(list.Get(i).Message())))
			}

		case fd.Kind() == protoreflect.MessageKind:
			trimMessage(value.Message())
		}

		return true
	})

	return m
}

func trimString(v protoreflect.Value) protoreflect.Value {
	return protoreflect.ValueOfString(strings.TrimSpace(v.String()))
}
