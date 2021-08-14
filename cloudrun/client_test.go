package cloudrun

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAuthClientLocal(t *testing.T) {
	if os.Getenv("CLOUDRUN_AUTH_TESTS") != "true" {
		t.Skip("Set CLOUDRUN_AUTH_TESTS=true to run this test")
	}

	client := NewAuthenticatedHTTPClient()

	req, err := http.NewRequest(http.MethodGet, "https://pkg-cloudrun-v2-itislcc66a-ew.a.run.app", nil)
	require.NoError(t, err)
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, resp.StatusCode, http.StatusOK)
}

func TestAuthClientServiceAccount(t *testing.T) {
	if os.Getenv("CLOUDRUN_AUTH_TESTS") != "true" {
		t.Skip("Set CLOUDRUN_AUTH_TESTS=true to run this test")
	}

	wd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", filepath.Join(wd, "testdata", "service-account.json")))

	client := NewAuthenticatedHTTPClient()

	req, err := http.NewRequest(http.MethodGet, "https://pkg-cloudrun-v2-itislcc66a-ew.a.run.app", nil)
	require.NoError(t, err)
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, resp.StatusCode, http.StatusOK)
}
