package cloudrun

import (
	"libs.altipla.consulting/hosting"
)

func Platform() hosting.Platform {
	return new(crplatform)
}

type crplatform struct{}

func (p *crplatform) Init() error {
	return nil
}
