package cloudrun

import (
	"context"
	"os"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"libs.altipla.consulting/env"
	"libs.altipla.consulting/errors"
	"libs.altipla.consulting/sentry"
)

func grpcUnaryErrorLogger() grpc.UnaryServerInterceptor {
	client := sentry.NewClient(os.Getenv("SENTRY_DSN"))

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if client != nil {
			defer client.ReportPanics(ctx)
		}

		ctx = sentry.WithContextRPC(ctx, env.ServiceName(), info.FullMethod)

		ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()

		resp, err := handler(ctx, req)
		if err != nil {
			logError(ctx, client, info.FullMethod, err)
		}

		return resp, err
	}
}

func logError(ctx context.Context, client *sentry.Client, method string, err error) {
	if env.IsLocal() {
		log.Println(errors.Stack(err))
	}

	grpcerr, ok := status.FromError(err)
	if ok {
		// Always log the GRPC errors.
		log.WithFields(log.Fields{
			"code":    grpcerr.Code().String(),
			"message": grpcerr.Message(),
			"method":  method,
		}).Error("GRPC call failed")

		// Do not notify those status codes.
		switch grpcerr.Code() {
		case codes.InvalidArgument, codes.NotFound, codes.AlreadyExists, codes.FailedPrecondition, codes.Aborted, codes.Unimplemented, codes.Canceled, codes.Unauthenticated, codes.ResourceExhausted:
			return
		}
	} else {
		log.WithFields(errors.LogFields(err)).Error("Unknown error in GRPC call")
	}

	// Do not notify UTF-8 decoding errors.
	//
	// These kind of errors are common when receiving Envoy access logs or validating
	// HTML inputs in admin editors.
	if strings.HasPrefix(err.Error(), "rpc error: code = Internal desc = grpc: failed to unmarshal the received message") && strings.HasSuffix(err.Error(), "contains invalid UTF-8") {
		return
	}

	if client != nil {
		client.Report(ctx, err)
	}
}
