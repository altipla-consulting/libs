package platform

import (
	"libs.altipla.consulting/env"
	"libs.altipla.consulting/hosting"
	"libs.altipla.consulting/hosting/bare"
	"libs.altipla.consulting/hosting/cloudrun"
	"libs.altipla.consulting/hosting/kubernetes"
)

func DetectFromEnv() hosting.Platform {
	if env.IsCloudRun() {
		return cloudrun.Platform()
	}
	if env.IsKubernetes() {
		return kubernetes.Platform()
	}
	return bare.Platform()
}
