package doris

import (
	"context"
	"os"
	"time"

	"github.com/bufbuild/connect-go"
	log "github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"

	"libs.altipla.consulting/env"
	"libs.altipla.consulting/errors"
	"libs.altipla.consulting/sentry"
)

func ServerInterceptors() []connect.Interceptor {
	return []connect.Interceptor{
		serverOnlyInterceptor(),
		genericTimeoutInterceptor(),
		trimRequestsInterceptor(),
		sentryLoggerInterceptor(),
	}
}

func serverOnlyInterceptor() connect.Interceptor {
	return connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
		return connect.UnaryFunc(func(ctx context.Context, in connect.AnyRequest) (connect.AnyResponse, error) {
			if in.Spec().IsClient {
				panic("do not configure server interceptors on a client instance")
			}
			return next(ctx, in)
		})
	})
}

func genericTimeoutInterceptor() connect.Interceptor {
	return connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
		return connect.UnaryFunc(func(ctx context.Context, in connect.AnyRequest) (connect.AnyResponse, error) {
			ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
			defer cancel()
			return next(ctx, in)
		})
	})
}

func trimRequestsInterceptor() connect.Interceptor {
	return connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
		return connect.UnaryFunc(func(ctx context.Context, in connect.AnyRequest) (connect.AnyResponse, error) {
			trimMessage(in.Any().(proto.Message).ProtoReflect())
			return next(ctx, in)
		})
	})
}

func sentryLoggerInterceptor() connect.Interceptor {
	client := sentry.NewClient(os.Getenv("SENTRY_DSN"))

	return connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
		return connect.UnaryFunc(func(ctx context.Context, in connect.AnyRequest) (connect.AnyResponse, error) {
			defer client.ReportPanics(ctx)

			ctx = sentry.WithContextRPC(ctx, env.ServiceName(), in.Spec().Procedure)

			reply, err := next(ctx, in)
			if err != nil {
				logError(ctx, client, in.Spec().Procedure, err)
			}
			return reply, err
		})
	})
}

func logError(ctx context.Context, client *sentry.Client, method string, err error) {
	if env.IsLocal() {
		log.Println(errors.Stack(err))
	}

	if connecterr := new(connect.Error); errors.As(err, &connecterr) {
		// Always log the Connect errors.
		log.WithFields(log.Fields{
			"code":    connecterr.Code().String(),
			"message": connecterr.Message(),
			"method":  method,
		}).Error("Connect call failed")

		// Do not notify those status codes.
		switch connecterr.Code() {
		case connect.CodeInvalidArgument, connect.CodeNotFound, connect.CodeAlreadyExists, connect.CodeFailedPrecondition, connect.CodeAborted, connect.CodeUnimplemented, connect.CodeCanceled, connect.CodeUnauthenticated, connect.CodeResourceExhausted:
			return
		}
	} else {
		log.WithFields(errors.LogFields(err)).Error("Unknown error in Connect call")
	}

	// Do not notify disconnections from the client.
	if ctx.Err() == context.Canceled {
		return
	}

	client.Report(ctx, err)
}
