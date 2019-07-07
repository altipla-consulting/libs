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

func init() {
	if IsLocal() {
		log.SetFormatter(&log.TextFormatter{
			ForceColors: true,
		})
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetFormatter(new(stackdriverFormatter))
	}
}

type stackdriverFormatter struct {
	log.JSONFormatter
}

func (f *stackdriverFormatter) Format(entry *log.Entry) ([]byte, error) {
	switch entry.Level {
	// EMERGENCY (800) One or more systems are unusable.
	case log.PanicLevel:
		entry.Data["severity"] = 800

	// ALERT (700) A person must take an action immediately.
	case log.FatalLevel:
		entry.Data["severity"] = 700

	// CRITICAL (600) Critical events cause more severe problems or outages.
	// No equivalent in logrus.

	// ERROR (500) Error events are likely to cause problems.
	case log.ErrorLevel:
		entry.Data["severity"] = 500

	// WARNING (400) Warning events might cause problems.
	case log.WarnLevel:
		entry.Data["severity"] = 400

	// NOTICE (300) Normal but significant events, such as start up, shut down, or a configuration change.
	// No equivalent in logrus.

	// INFO (200) Routine information, such as ongoing status or performance.
	case log.InfoLevel:
		entry.Data["severity"] = 200

	// DEBUG (100) Debug or trace information.
	case log.DebugLevel:
		entry.Data["severity"] = 100

	// DEFAULT (0) The log entry has no assigned severity level.
	case log.TraceLevel:
		entry.Data["severity"] = 0
	}

	return f.JSONFormatter.Format(entry)
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
