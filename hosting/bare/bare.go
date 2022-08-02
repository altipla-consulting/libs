package bare

import (
	"context"
	"fmt"
	"net/http"

	"libs.altipla.consulting/env"
	"libs.altipla.consulting/errors"
	"libs.altipla.consulting/hosting"
	"libs.altipla.consulting/routing"
)

func Platform() hosting.Platform {
	return new(barePlatform)
}

type barePlatform struct {
	internal *http.Server
}

func (platform *barePlatform) Init() error {
	r := routing.NewServer(routing.WithLogrus())
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) error {
		fmt.Fprintf(w, "%s is ok\n", env.ServiceName())
		return nil
	})

	platform.internal = &http.Server{
		Addr:    ":8000",
		Handler: r,
	}
	go platform.internal.ListenAndServe()

	return nil
}

func (platform *barePlatform) Shutdown(ctx context.Context) error {
	if err := platform.internal.Shutdown(ctx); err != nil {
		return errors.Trace(err)
	}

	return nil
}
