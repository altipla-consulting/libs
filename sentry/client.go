package sentry

import (
	"context"
	"net/http"
	"os"

	"github.com/getsentry/raven-go"
	log "github.com/sirupsen/logrus"

	"libs.altipla.consulting/errors"
)

// Client wraps a Sentry connection.
type Client struct {
	dsn string
}

// NewClient opens a new connection to the Sentry report API.
func NewClient(dsn string) *Client {
	return &Client{
		dsn: dsn,
	}
}

// ReportInternal reports an error not linked to a HTTP request.
func (client *Client) ReportInternal(ctx context.Context, appErr error) {
	client.report(ctx, appErr, nil)
}

// ReportRequest reports an error linked to a HTTP request.
func (client *Client) ReportRequest(appErr error, r *http.Request) {
	client.report(r.Context(), appErr, r)
}

func (client *Client) report(ctx context.Context, appErr error, r *http.Request) {
	go func() {
		stacktrace := new(raven.Stacktrace)

		for i, stack := range errors.Frames(appErr) {
			if i > 0 {
				stacktrace.Frames = append(stacktrace.Frames, &raven.StacktraceFrame{
					Filename: "------",
				})
			}

			for _, frame := range stack {
				stacktrace.Frames = append(stacktrace.Frames, &raven.StacktraceFrame{
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

		client, err := raven.New(client.dsn)
		if err != nil {
			log.WithField("error", err.Error()).Error("Cannot create Sentry client")
			return
		}
		client.SetRelease(os.Getenv("VERSION"))

		interfaces := []raven.Interface{
			&ravenException{
				Stacktrace: stacktrace,
				Module:     "backend",
				Value:      appErr.Error(),
				Type:       appErr.Error(),
			},
		}

		info := FromContext(ctx)
		if info != nil {
			interfaces = append(interfaces, &ravenBreadcrumbs{
				Values: info.breadcrumbs,
			})
		}

		if r != nil {
			interfaces = append(interfaces, raven.NewHttp(r))
		}

		packet := raven.NewPacket(appErr.Error(), interfaces...)
		if info != nil && info.rpcMethod != "" {
			packet.AddTags(map[string]string{
				"rpc_service": info.rpcService,
				"rpc_method":  info.rpcMethod,
			})
		}

		eventID, ch := client.Capture(packet, nil)
		<-ch
		log.WithField("eventID", eventID).Info("Error logged to sentry")
	}()
}
