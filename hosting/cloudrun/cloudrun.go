package cloudrun

import (
	"context"

	"libs.altipla.consulting/hosting"
)

func Platform() hosting.Platform {
	return new(cloudRunPlatform)
}

type cloudRunPlatform struct{}

func (platform *cloudRunPlatform) Init() error {
	return nil
}

func (platform *cloudRunPlatform) Shutdown(ctx context.Context) error {
	return nil
}
