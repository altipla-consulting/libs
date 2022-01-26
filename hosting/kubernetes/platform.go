package kubernetes

import (
	"libs.altipla.consulting/hosting"
)

func Platform() hosting.Platform {
	return new(k8splatform)
}

type k8splatform struct{}

func (p *k8splatform) Init() error {
	return nil
}
