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

// Version returns the environment variable VERSION. In development it should be empty.
// In production it should be set accordingly; it may be for example the container hash.
func Version() string {
	return os.Getenv("VERSION")
}
