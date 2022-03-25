package kubernetes

import (
	"context"
	"fmt"
	"net/http"
	"time"

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
	if err := platform.internal.Shutdown(ctx); err != nil {
		return errors.Trace(err)
	}

	// Wait 5 seconds before shutting down the rest of servers so Kubernetes has
	// enough time to redirect traffic to other instances.
	time.Sleep(5 * time.Second)

	return nil
}
