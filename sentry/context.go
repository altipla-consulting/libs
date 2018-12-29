package sentry

import (
	"context"
)

type key int

var keySentry key = 1

// Sentry accumulates info through out the whole request to send them in case
// an error is reported.
type Sentry struct {
	breadcrumbs           []*ravenBreadcrumb
	rpcService, rpcMethod string
}

// FromContext returns the Sentry instance stored in the context. If no instance
// was created it will return nil.
func FromContext(ctx context.Context) *Sentry {
	value := ctx.Value(keySentry)
	if value == nil {
		return nil
	}

	return value.(*Sentry)
}

// WithContext stores a new instance of Sentry in the context and returns the
// new generated context that you should use everywhere.
func WithContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, keySentry, new(Sentry))
}

// WithContextRPC stores a new instance of Sentry in the context and returns the
// new generated context that you should use everywhere annotating it with the
// RPC service name and RPC method name. Used mostly for GRPC calls.
func WithContextRPC(ctx context.Context, rpcService, rpcMethod string) context.Context {
	sentry := &Sentry{
		rpcService: rpcService,
		rpcMethod:  rpcMethod,
	}
	return context.WithValue(ctx, keySentry, sentry)
}
