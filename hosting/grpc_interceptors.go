package hosting

import (
	"context"
	"os"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	"libs.altipla.consulting/env"
	"libs.altipla.consulting/errors"
	"libs.altipla.consulting/sentry"
)

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

func grpcUnaryErrorLogger() grpc.UnaryServerInterceptor {
	client := sentry.NewClient(os.Getenv("SENTRY_DSN"))

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		defer client.ReportPanics(ctx)

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

	// Do not notify disconnections from the client.
	if ctx.Err() == context.Canceled {
		return
	}

	client.Report(ctx, err)
}
