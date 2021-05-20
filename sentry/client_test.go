package sentry_test

import (
	"context"
	"net/http"
	"os"
	"testing"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"libs.altipla.consulting/errors"
	"libs.altipla.consulting/sentry"
)

func TestReportPanic(t *testing.T) {
	if os.Getenv("SENTRY_DSN") == "" {
		t.Skip("Skipping sentry real tests without SENTRY_DSN=foo env variable")
	}
	defer time.Sleep(3 * time.Second)

	client := sentry.NewClient(os.Getenv("SENTRY_DSN"))
	defer client.ReportPanics(context.Background())

	panic("foo")
}

func TestIgnoreAbortError(t *testing.T) {
	if os.Getenv("SENTRY_DSN") == "" {
		t.Skip("Skipping sentry real tests without SENTRY_DSN=foo env variable")
	}
	defer time.Sleep(3 * time.Second)

	client := sentry.NewClient(os.Getenv("SENTRY_DSN"))
	defer client.ReportPanics(context.Background())

	panic(http.ErrAbortHandler)
}

func TestGRPCError(t *testing.T) {
	if os.Getenv("SENTRY_DSN") == "" {
		t.Skip("Skipping sentry real tests without SENTRY_DSN=foo env variable")
	}
	defer time.Sleep(3 * time.Second)

	client := sentry.NewClient(os.Getenv("SENTRY_DSN"))

	client.Report(context.Background(), errors.Trace(status.Errorf(codes.NotFound, "not found example")))
}
