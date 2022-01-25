package sentry

import (
	"context"
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/getsentry/sentry-go"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc/status"

	"libs.altipla.consulting/env"
	"libs.altipla.consulting/errors"
)

const (
	// LevelCritical is a critical breadcrumb.
	LevelFatal = sentry.LevelFatal

	// LevelWarning is a warning breadcrumb.
	LevelWarning = sentry.LevelWarning

	// LevelError is an error breadcrumb.
	LevelError = sentry.LevelError

	// LevelInfo is an info breadcrumb.
	LevelInfo = sentry.LevelInfo

	// LevelDebug is a debug breadcrumb.
	LevelDebug = sentry.LevelDebug
)

// Client wraps a Sentry connection.
type Client struct {
	hub *sentry.Hub
}

// NewClient opens a new connection to the Sentry report API.
func NewClient(dsn string) *Client {
	if dsn == "" {
		return nil
	}

	client, err := sentry.NewClient(sentry.ClientOptions{
		Dsn:     dsn,
		Release: env.Version(),
	})
	if err != nil {
		panic(err)
	}

	return &Client{hub: sentry.NewHub(client, sentry.NewScope())}
}

// Report reports an error to Sentry.
func (client *Client) Report(ctx context.Context, appErr error) {
	if client == nil {
		return
	}
	client.sendReport(ctx, appErr, nil)
}

// ReportRequest reports an error linked to a HTTP request.
func (client *Client) ReportRequest(r *http.Request, appErr error) {
	if client == nil {
		return
	}
	client.sendReport(r.Context(), appErr, r)
}

// ReportPanics detects panics in the rest of the body of the function and
// reports it if one occurs.
func (client *Client) ReportPanics(ctx context.Context) {
	if client == nil {
		return
	}
	client.ReportPanic(ctx, recover())
}

// ReportPanic sends a panic correctly formated to the server if the argument
// is not nil.
func (client *Client) ReportPanic(ctx context.Context, panicErr interface{}) {
	if client == nil || panicErr == nil {
		return
	}
	rec := errors.Recover(panicErr)

	// Ignoramos el error causado por abortar un handler.
	if errors.Is(rec, http.ErrAbortHandler) {
		panic(panicErr)
	}

	log.WithField("error", rec.Error()).Error("Panic recovered")
	client.sendReportPanic(ctx, rec, string(debug.Stack()), nil)
}

// ReportPanicsRequest detects pancis in the body of the function and reports them
// linked to a HTTP request.
func (client *Client) ReportPanicsRequest(r *http.Request) {
	if client == nil {
		return
	}
	if rec := errors.Recover(recover()); rec != nil {
		log.WithField("error", rec.Error()).Error("Panic recovered")
		client.sendReportPanic(r.Context(), rec, string(debug.Stack()), r)
	}
}

func (client *Client) sendReport(ctx context.Context, appErr error, r *http.Request) {
	go func() {
		stacktrace := new(sentry.Stacktrace)
		for i, stack := range errors.Frames(appErr) {
			if i > 0 {
				stacktrace.Frames = append(stacktrace.Frames, sentry.Frame{
					Filename: "------",
				})
			}

			for _, frame := range stack {
				stacktrace.Frames = append(stacktrace.Frames, sentry.Frame{
					Filename:    frame.File,
					Function:    frame.Function,
					Lineno:      frame.Line,
					ContextLine: frame.Reason,
				})
			}
		}

		// Invert frames to show them in the correct order in the Sentry UI
		for i, j := 0, len(stacktrace.Frames)-1; i < j; i, j = i+1, j-1 {
			stacktrace.Frames[i], stacktrace.Frames[j] = stacktrace.Frames[j], stacktrace.Frames[i]
		}

		event := sentry.NewEvent()

		// Do not send empty stacktraces, it's an error.
		if len(stacktrace.Frames) != 0 {
			event.Exception = []sentry.Exception{
				{
					Stacktrace: stacktrace,
					Module:     "backend",
					Value:      appErr.Error(),
					Type:       appErr.Error(),
				},
			}
		}

		info := FromContext(ctx)
		if info != nil {
			event.Breadcrumbs = info.breadcrumbs

			if info.rpcMethod != "" {
				event.Extra["rpc_service"] = info.rpcService
				event.Extra["rpc_method"] = info.rpcMethod
			}
		}

		if r != nil {
			event.Request = sentry.NewRequest(r)
		}

		if st, ok := status.FromError(appErr); ok {
			event.Tags["grpc_code"] = fmt.Sprintf("%v", st.Code())
			event.Extra["grpc_message"] = st.Message()
		}

		eventID := client.hub.CaptureEvent(event)
		log.WithField("eventID", eventID).Info("Error sent to Sentry")
	}()
}

func (client *Client) sendReportPanic(ctx context.Context, appErr error, message string, r *http.Request) {
	go func() {
		event := sentry.NewEvent()

		event.Message = message
		event.Exception = []sentry.Exception{
			{
				Type:   appErr.Error(),
				Module: "backend",
			},
		}

		info := FromContext(ctx)
		if info != nil {
			event.Breadcrumbs = info.breadcrumbs

			if info.rpcMethod != "" {
				event.Extra["rpc_service"] = info.rpcService
				event.Extra["rpc_method"] = info.rpcMethod
			}
		}

		if r != nil {
			event.Request = sentry.NewRequest(r)
		}

		eventID := client.hub.CaptureEvent(event)
		log.WithField("event-id", eventID).Info("Error sent to Sentry")
	}()
}

// LogBreadcrumb logs a new breadcrumb in the Sentry instance of the context.
func LogBreadcrumb(ctx context.Context, level sentry.Level, category, message string) {
	info := FromContext(ctx)
	if info == nil {
		return
	}

	info.breadcrumbs = append(info.breadcrumbs, &sentry.Breadcrumb{
		Timestamp: time.Now(),
		Type:      "default",
		Message:   message,
		Category:  category,
		Level:     level,
	})
}
