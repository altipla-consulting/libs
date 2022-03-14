package cloudrun

import (
	"context"

	"libs.altipla.consulting/hosting"
)

func Platform() hosting.Platform {
	return new(crplatform)
}

type crplatform struct{}

func (platform *crplatform) Init() error {
	return nil
}

func (platform *crplatform) Shutdown(ctx context.Context) error {
	return nil
}
