package doris

import "libs.altipla.consulting/routing"

// Option of a server.
type Option func(cnf *config)

type config struct {
	http     []routing.ServerOption
	profiler bool
}

// WithRoutingOptions configures web server options.
func WithRoutingOptions(opts ...routing.ServerOption) Option {
	return func(cnf *config) {
		cnf.http = append(cnf.http, opts...)
	}
}

// WithProfiler enables the Google Cloud Profiler for the application.
func WithProfiler() Option {
	return func(cnf *config) {
		cnf.profiler = true
	}
}
