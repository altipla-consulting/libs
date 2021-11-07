package env

import (
	"io/ioutil"
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

// IsAzureFunction detects if we are running inside an Azure Function app.
func IsAzureFunction() bool {
	return os.Getenv("APPSETTING_WEBSITE_SITE_NAME") != ""
}

// Version returns the application version. In supported environment it may extract
// the info from files or environment variables. Otherwise it will use the env variable
// VERSION that should be set manually to the desired value.
func Version() string {
	// Cloud Run.
	if e := os.Getenv("K_REVISION"); e != "" {
		return e
	}

	// Azure Function has a Kubu file one level up of the working directory.
	if IsAzureFunction() {
		v, err := ioutil.ReadFile("../deployments/active")
		if err == nil {
			return string(v)
		} else if !os.IsNotExist(err) {
			panic(err)
		}
	}

	// Generic settings we have to set as custom options in each platform.
	return os.Getenv("VERSION")
}

// ServiceName returns the version of the application from the environment-dependent
// variables or the hostname if none is available.
func ServiceName() string {
	// Cloud Run.
	if v := os.Getenv("K_SERVICE"); v != "" {
		return v
	}

	// Azure Functions.
	if v := os.Getenv("APPSETTING_WEBSITE_SITE_NAME"); v != "" {
		return v
	}

	// Last option is a hostname which should always exist, though is not very useful usually.
	v, err := os.Hostname()
	if err != nil {
		panic(err)
	}
	return v
}
