package services

import (
	"context"
	"strings"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"libs.altipla.consulting/errors"
	"libs.altipla.consulting/sentry"
)

// Endpoint is a simple string with the host and port of the remote GRPC
// service. We use a custom type to avoid using grpc.Dial without noticing the bug.
//
// This needs a "discovery" package with the full list of remote addresses that
// use this type instead of string and never using the direct address. That way
// if you use grpc.Dial it will report the compilation error inmediatly.
type Endpoint string

// Dial helps to open a connection to a remote GRPC server.
func Dial(target Endpoint, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	return grpc.Dial(string(target), opts...)
}

func grpcUnaryErrorLogger(serviceName, dsn string) grpc.UnaryServerInterceptor {
	var client *sentry.Client
	if dsn != "" {
		client = sentry.NewClient(dsn)
	}

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if client != nil {
			defer client.ReportPanics(ctx)
		}

		ctx = sentry.WithContextRPC(ctx, serviceName, info.FullMethod)

		resp, err := handler(ctx, req)
		if err != nil {
			logError(ctx, client, err)
		}

		return resp, err
	}
}

type wrappedStream struct {
	grpc.ServerStream
	serviceName, method string
}

func (stream *wrappedStream) Context() context.Context {
	ctx := stream.ServerStream.Context()
	ctx = sentry.WithContextRPC(ctx, stream.serviceName, stream.method)

	return ctx
}

func grpcStreamErrorLogger(serviceName, dsn string) grpc.StreamServerInterceptor {
	var client *sentry.Client
	if dsn != "" {
		client = sentry.NewClient(dsn)
	}

	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		wrapped := &wrappedStream{
			ServerStream: stream,
			serviceName:  serviceName,
			method:       info.FullMethod,
		}

		if client != nil {
			defer client.ReportPanics(wrapped.Context())
		}

		err := handler(srv, wrapped)
		if err != nil {
			logError(wrapped.Context(), client, err)
		}

		return errors.Trace(err)
	}
}

func logError(ctx context.Context, client *sentry.Client, err error) {
	grpcerr, ok := status.FromError(err)
	if ok {
		// Always log the GRPC errors.
		log.WithFields(log.Fields{
			"code":    grpcerr.Code().String(),
			"message": grpcerr.Message(),
		}).Error("GRPC call failed")

		// Do not notify those status codes.
		switch grpcerr.Code() {
		case codes.InvalidArgument, codes.NotFound, codes.AlreadyExists, codes.FailedPrecondition, codes.Aborted, codes.Unimplemented, codes.Canceled, codes.Unauthenticated:
			return
		}
	} else {
		log.WithFields(errors.LogFields(err)).Error("Unknown error in GRPC call")
	}

	// Do not notify UTF-8 decoding errors.
	//
	// These kind of errors are common when receiving Envoy access logs or validating
	// HTML inputs in admin editors.
	if strings.HasPrefix(err.Error(), "rpc error: code = Internal desc = grpc: failed to unmarshal the received message proto: field") && strings.HasSuffix(err.Error(), "contains invalid UTF-8") {
		return
	}

	if client != nil {
		client.ReportInternal(ctx, err)
	}
}
