package sentry

import (
	"context"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/getsentry/raven-go"
	log "github.com/sirupsen/logrus"
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

type jujuStacktracer interface {
	StackTrace() []string
}

func (client *Client) report(ctx context.Context, appErr error, r *http.Request) {
	go func() {
		stacktrace := new(raven.Stacktrace)

		jujuErr, ok := appErr.(jujuStacktracer)
		if ok {
			for _, entry := range jujuErr.StackTrace() {
				parts := strings.Split(entry, ":")
				if len(parts) > 2 {
					n, err := strconv.ParseInt(parts[1], 10, 64)
					if err == nil {
						stacktrace.Frames = append(stacktrace.Frames, &raven.StacktraceFrame{
							Filename:    parts[0],
							Lineno:      int(n),
							ContextLine: entry,
						})
						continue
					}
				}

				// Fallback to avoid erroring out here if no location is found
				stacktrace.Frames = append(stacktrace.Frames, &raven.StacktraceFrame{
					Filename:    entry,
					ContextLine: entry,
				})
			}

			// Invert frames to show them in the correct order in the Sentry UI
			for i, j := 0, len(stacktrace.Frames)-1; i < j; i, j = i+1, j-1 {
				stacktrace.Frames[i], stacktrace.Frames[j] = stacktrace.Frames[j], stacktrace.Frames[i]
			}
		}

		client, err := raven.New(client.dsn)
		if err != nil {
			log.WithField("error", err).Error("Cannot create client")
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
