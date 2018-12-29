package services

import (
	"os"
)

// IsLocal returns true if we are running inside a local debug environment instead
// of a production Kubernetes container. It dependes on Version() working correctly.
func IsLocal() bool {
	return Version() == ""
}

// Version returns the environment variable VERSION. In development it should be empty.
// In production it should be set accordingly; it may be for example the container hash.
func Version() string {
	return os.Getenv("VERSION")
}
