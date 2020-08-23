package env

import (
	"os"
)

// IsLocal returns true if we are running inside a local debug environment instead
// of a production Kubernetes container. It dependes on Version() working correctly.
func IsLocal() bool {
	return Version() == ""
}

// IsJenkins detects if we are running as a step of a Jenkins build.
func IsJenkins() bool {
	return os.Getenv("BUILD_ID") != ""
}

// IsCloudRun detects if we are running inside a Cloud Run app.
func IsCloudRun() bool {
	return os.Getenv("K_CONFIGURATION") != ""
}

// Version returns the environment variable K_REVISION or VERSION. In development
// it should be empty. In production it should be set accordingly; it may be for
// example the container hash.
func Version() string {
	if e := os.Getenv("K_REVISION"); e != "" {
		return e
	}
	return os.Getenv("VERSION")
}

// ServiceName returns the K_SERVICE environment variable or the hostname if not
// available as it is common during development.
func ServiceName() string {
	if v := os.Getenv("K_SERVICE"); v != "" {
		return v
	}
	v, err := os.Hostname()
	if err != nil {
		panic(err)
	}
	return v
}
