package sentry

import (
	"context"
	"time"

	"github.com/getsentry/raven-go"
)

type ravenBreadcrumbs struct {
	Values []*ravenBreadcrumb `json:"values,omitempty"`
}

// Class implements the raven.Interface type.
func (b *ravenBreadcrumbs) Class() string {
	return "breadcrumbs"
}

type ravenBreadcrumb struct {
	Timestamp time.Time         `json:"timestamp"`
	Type      string            `json:"type"`
	Message   string            `json:"message"`
	Data      map[string]string `json:"data,omitempty"`
	Category  string            `json:"category"`
	Level     Level             `json:"level"`
}

// Level of a breadcrumb.
type Level string

const (
	// LevelCritical is a critical breadcrumb.
	LevelCritical = Level("critical")

	// LevelWarning is a warning breadcrumb.
	LevelWarning = Level("warning")

	// LevelError is an error breadcrumb.
	LevelError = Level("error")

	// LevelInfo is an info breadcrumb.
	LevelInfo = Level("info")

	// LevelDebug is a debug breadcrumb.
	LevelDebug = Level("debug")
)

// LogBreadcrumb logs a new breadcrumb in the Sentry instance of the context.
func LogBreadcrumb(ctx context.Context, level Level, category, message string) {
	info := FromContext(ctx)
	if info == nil {
		return
	}

	info.breadcrumbs = append(info.breadcrumbs, &ravenBreadcrumb{
		Timestamp: time.Now(),
		Type:      "default",
		Message:   message,
		Category:  category,
		Level:     level,
	})
}

type ravenException struct {
	Value      string            `json:"value"`
	Module     string            `json:"module"`
	Stacktrace *raven.Stacktrace `json:"stacktrace"`
	Type       string            `json:"type"`
}

// Class implements the raven.Interface type.
func (e *ravenException) Class() string {
	return "exception"
}

type ravenRPC struct {
}
