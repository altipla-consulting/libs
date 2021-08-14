package cloudrun

import (
	"google.golang.org/grpc"
	"libs.altipla.consulting/routing"
)

// Option of a Cloud Run server.
type Option func(cnf *config)

type config struct {
	http              []routing.ServerOption
	profiler          bool
	grpc              []grpc.ServerOption
	unaryInterceptors []grpc.UnaryServerInterceptor
	queues            func(*routing.Server)
	cors              []string
}

// WithRoutingOptions configures web server options. Only valid with the Web() constructor.
func WithRoutingOptions(opts ...routing.ServerOption) Option {
	return func(cnf *config) {
		cnf.http = append(cnf.http, opts...)
	}
}

// WithGRPCOptions configures GRPC server options. Only valid with the GRPC() constructor.
func WithGRPCOptions(opts ...grpc.ServerOption) Option {
	return func(cnf *config) {
		cnf.grpc = append(cnf.grpc, opts...)
	}
}

// WithProfiler enables the Google Cloud Profiler for the application.
func WithProfiler() Option {
	return func(cnf *config) {
		cnf.profiler = true
	}
}

// WithQueues initializes and configures the queues.
func WithQueues(fn func(*routing.Server)) Option {
	return func(cnf *config) {
		cnf.queues = fn
	}
}

func WithAuth(auth Authorizer) Option {
	return func(cnf *config) {
		cnf.unaryInterceptors = append(cnf.unaryInterceptors, auth.GRPCInterceptor())
	}
}

// WithCORS authorizes domains to access resources from the specified JS domains.
// Wildcards are supported as prefixes and suffixes.
//
// For example:
//     foo.altipla.consulting
//     *.foo.altipla.consulting
//     altipla.consulting/foo/*
func WithCORS(domains ...string) Option {
	return func(cnf *config) {
		cnf.cors = append(cnf.cors, domains...)
	}
}

type Authorizer interface {
	GRPCInterceptor() grpc.UnaryServerInterceptor

	// TODO(alberto): Eliminar cuando services/v2 desimplemente el metodo
	CheckIDToken(audience, subject string, handler routing.Handler) routing.Handler
}
