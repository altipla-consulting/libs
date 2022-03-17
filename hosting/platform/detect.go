package platform

import (
	"libs.altipla.consulting/env"
	"libs.altipla.consulting/hosting"
	"libs.altipla.consulting/hosting/cloudrun"
	"libs.altipla.consulting/hosting/kubernetes"
)

func DetectFromEnv() hosting.Platform {
	if env.IsCloudRun() {
		return cloudrun.Platform()
	}
	return kubernetes.Platform()
}
