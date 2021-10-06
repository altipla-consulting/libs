package routing

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func fakeRequest(t *testing.T, server *Server, req *http.Request) (*http.Response, string) {
	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)
	resp := w.Result()
	body, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp, string(body)
}

func TestSimpleRoute(t *testing.T) {
	server := NewServer(WithLogrus())
	server.Get("/test", func(w http.ResponseWriter, r *http.Request) error {
		fmt.Fprint(w, "ok")
		return nil
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	resp, body := fakeRequest(t, server, req)
	require.Equal(t, resp.StatusCode, http.StatusOK)

	require.Equal(t, body, "ok")
}

func Test404(t *testing.T) {
	server := NewServer(WithLogrus())

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	resp, _ := fakeRequest(t, server, req)
	require.Equal(t, resp.StatusCode, http.StatusNotFound)
}

func TestDomains(t *testing.T) {
	server := NewServer(WithLogrus())
	foo := server.Domain("www.foo.com")
	foo.Get("/test", func(w http.ResponseWriter, r *http.Request) error {
		fmt.Fprint(w, "foo")
		return nil
	})
	bar := server.Domain("www.bar.com")
	bar.Get("/test", func(w http.ResponseWriter, r *http.Request) error {
		fmt.Fprint(w, "bar")
		return nil
	})

	req := httptest.NewRequest(http.MethodGet, "http://www.foo.com/test", nil)
	resp, body := fakeRequest(t, server, req)
	require.Equal(t, resp.StatusCode, http.StatusOK)
	require.Equal(t, body, "foo")

	req = httptest.NewRequest(http.MethodGet, "http://www.bar.com/test", nil)
	resp, body = fakeRequest(t, server, req)
	require.Equal(t, resp.StatusCode, http.StatusOK)
	require.Equal(t, body, "bar")

	req = httptest.NewRequest(http.MethodGet, "http://www.baz.com/test", nil)
	resp, _ = fakeRequest(t, server, req)
	require.Equal(t, resp.StatusCode, http.StatusNotFound)
}

func TestDomainsWithPathPrefix(t *testing.T) {
	server := NewServer(WithLogrus())
	domain := server.Domain("www.foo.com").PathPrefix("/admin")
	domain.Get("/test", func(w http.ResponseWriter, r *http.Request) error {
		fmt.Fprint(w, "ok")
		return nil
	})

	req := httptest.NewRequest(http.MethodGet, "http://www.foo.com/admin/test", nil)
	resp, body := fakeRequest(t, server, req)
	require.Equal(t, resp.StatusCode, http.StatusOK)
	require.Equal(t, body, "ok")

	req = httptest.NewRequest(http.MethodGet, "http://www.foo.com/test", nil)
	resp, _ = fakeRequest(t, server, req)
	require.Equal(t, resp.StatusCode, http.StatusNotFound)

	req = httptest.NewRequest(http.MethodGet, "http://www.bar.com/admin/test", nil)
	resp, _ = fakeRequest(t, server, req)
	require.Equal(t, resp.StatusCode, http.StatusNotFound)
}

func TestPathPrefixChain(t *testing.T) {
	server := NewServer(WithLogrus())
	domain := server.PathPrefix("/admin").PathPrefix("/sub")
	domain.Get("/test", func(w http.ResponseWriter, r *http.Request) error {
		fmt.Fprint(w, "ok")
		return nil
	})

	req := httptest.NewRequest(http.MethodGet, "/admin/sub/test", nil)
	resp, body := fakeRequest(t, server, req)
	require.Equal(t, resp.StatusCode, http.StatusOK)
	require.Equal(t, body, "ok")

	req = httptest.NewRequest(http.MethodGet, "/test", nil)
	resp, _ = fakeRequest(t, server, req)
	require.Equal(t, resp.StatusCode, http.StatusNotFound)

	req = httptest.NewRequest(http.MethodGet, "/admin/test", nil)
	resp, _ = fakeRequest(t, server, req)
	require.Equal(t, resp.StatusCode, http.StatusNotFound)
}

func TestPathPrefixTree(t *testing.T) {
	server := NewServer(WithLogrus())
	admin := server.PathPrefix("/admin")
	admin.Get("/direct", func(w http.ResponseWriter, r *http.Request) error {
		fmt.Fprint(w, "direct")
		return nil
	})
	sub := admin.PathPrefix("/sub")
	sub.Get("/test", func(w http.ResponseWriter, r *http.Request) error {
		fmt.Fprint(w, "test")
		return nil
	})

	req := httptest.NewRequest(http.MethodGet, "/admin/direct", nil)
	resp, body := fakeRequest(t, server, req)
	require.Equal(t, resp.StatusCode, http.StatusOK)
	require.Equal(t, body, "direct")

	req = httptest.NewRequest(http.MethodGet, "/admin/sub/test", nil)
	resp, body = fakeRequest(t, server, req)
	require.Equal(t, resp.StatusCode, http.StatusOK)
	require.Equal(t, body, "test")

	req = httptest.NewRequest(http.MethodGet, "/test", nil)
	resp, _ = fakeRequest(t, server, req)
	require.Equal(t, resp.StatusCode, http.StatusNotFound)

	req = httptest.NewRequest(http.MethodGet, "/direct", nil)
	resp, _ = fakeRequest(t, server, req)
	require.Equal(t, resp.StatusCode, http.StatusNotFound)

	req = httptest.NewRequest(http.MethodGet, "/admin/test", nil)
	resp, _ = fakeRequest(t, server, req)
	require.Equal(t, resp.StatusCode, http.StatusNotFound)
}

func TestPathPrefixWithHome(t *testing.T) {
	server := NewServer(WithLogrus())
	admin := server.PathPrefix("/admin")
	admin.Get("", func(w http.ResponseWriter, r *http.Request) error {
		fmt.Fprint(w, "home")
		return nil
	})
	sub := admin.PathPrefix("/sub")
	sub.Get("/test", func(w http.ResponseWriter, r *http.Request) error {
		fmt.Fprint(w, "test")
		return nil
	})

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	resp, body := fakeRequest(t, server, req)
	require.Equal(t, resp.StatusCode, http.StatusOK)
	require.Equal(t, body, "home")

	req = httptest.NewRequest(http.MethodGet, "/admin/sub/test", nil)
	resp, body = fakeRequest(t, server, req)
	require.Equal(t, resp.StatusCode, http.StatusOK)
	require.Equal(t, body, "test")

	req = httptest.NewRequest(http.MethodGet, "/test", nil)
	resp, _ = fakeRequest(t, server, req)
	require.Equal(t, resp.StatusCode, http.StatusNotFound)

	req = httptest.NewRequest(http.MethodGet, "/admin/test", nil)
	resp, _ = fakeRequest(t, server, req)
	require.Equal(t, resp.StatusCode, http.StatusNotFound)
}

func TestDomainSameRouteAsRoot(t *testing.T) {
	server := NewServer(WithLogrus())
	domain := server.Domain("www.foo.com")
	domain.Get("/test", func(w http.ResponseWriter, r *http.Request) error {
		fmt.Fprint(w, "domain")
		return nil
	})
	server.Get("/test", func(w http.ResponseWriter, r *http.Request) error {
		fmt.Fprint(w, "root")
		return nil
	})

	req := httptest.NewRequest(http.MethodGet, "http://www.foo.com/test", nil)
	resp, body := fakeRequest(t, server, req)
	require.Equal(t, resp.StatusCode, http.StatusOK)
	require.Equal(t, body, "domain")

	req = httptest.NewRequest(http.MethodGet, "http://www.bar.com/test", nil)
	resp, body = fakeRequest(t, server, req)
	require.Equal(t, resp.StatusCode, http.StatusOK)
	require.Equal(t, body, "root")
}
