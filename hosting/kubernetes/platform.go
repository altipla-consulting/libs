package kubernetes

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
	return new(k8splatform)
}

type k8splatform struct {
	internal *http.Server
}

func (platform *k8splatform) Init() error {
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

func (platform *k8splatform) Shutdown(ctx context.Context) error {
	return errors.Trace(platform.internal.Shutdown(ctx))
}
