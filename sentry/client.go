package sentry

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/getsentry/sentry-go"
	log "github.com/sirupsen/logrus"

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
	client, err := sentry.NewClient(sentry.ClientOptions{
		Dsn:     dsn,
		Release: os.Getenv("VERSION"),
	})
	if err != nil {
		panic(err)
	}

	return &Client{hub: sentry.NewHub(client, sentry.NewScope())}
}

// ReportInternal reports an error not linked to a HTTP request.
func (client *Client) ReportInternal(ctx context.Context, appErr error) {
	client.Report(ctx, appErr)
}

// Report reports an error to Sentry.
func (client *Client) Report(ctx context.Context, appErr error) {
	client.report(ctx, appErr, nil)
}

// ReportRequest reports an error linked to a HTTP request.
func (client *Client) ReportRequest(appErr error, r *http.Request) {
	client.report(r.Context(), appErr, r)
}

// ReportPanics detects panics in the body of the function and reports them.
func (client *Client) ReportPanics(ctx context.Context) {
	if rec := recover(); rec != nil {
		appErr, ok := rec.(error)
		if !ok {
			appErr = errors.Errorf("panic: %v", rec)
		}
		client.report(ctx, appErr, nil)
	}
}

// ReportPanicsRequest detects pancis in the body of the function and reports them
// linked to a HTTP request.
func (client *Client) ReportPanicsRequest(r *http.Request) {
	if rec := recover(); rec != nil {
		appErr, ok := rec.(error)
		if !ok {
			appErr = errors.Errorf("panic: %v", rec)
		}
		client.report(r.Context(), appErr, r)
	}
}

func (client *Client) report(ctx context.Context, appErr error, r *http.Request) {
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
			event.Request = event.Request.FromHTTPRequest(r)
		}

		eventID := client.hub.CaptureEvent(event)
		log.WithField("eventID", eventID).Info("Error logged to sentry")
	}()
}

// LogBreadcrumb logs a new breadcrumb in the Sentry instance of the context.
func LogBreadcrumb(ctx context.Context, level sentry.Level, category, message string) {
	info := FromContext(ctx)
	if info == nil {
		return
	}

	info.breadcrumbs = append(info.breadcrumbs, &sentry.Breadcrumb{
		Timestamp: time.Now().Unix(),
		Type:      "default",
		Message:   message,
		Category:  category,
		Level:     level,
	})
}
