package rev

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/altipla-consulting/errors"

	"libs.altipla.consulting/env"
	"libs.altipla.consulting/loaders"
	"libs.altipla.consulting/routing"
)

var (
	assets     = map[string]string{}
	assetsLoad sync.Once
)

func Map(url string) (string, error) {
	fn := func() error {
		f, err := os.Open("rev-manifest.json")
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}

			return errors.Trace(err)
		}
		defer f.Close()

		return errors.Trace(json.NewDecoder(f).Decode(&assets))
	}
	if err := loaders.Lazy(&assetsLoad, fn); err != nil {
		return "", errors.Trace(err)
	}

	if m, ok := assets[url]; ok {
		return m, nil
	}
	return url, nil
}

func MapHandler(w http.ResponseWriter, r *http.Request) error {
	dest, err := Map(r.URL.Path)
	if err != nil {
		return errors.Trace(err)
	}
	if dest == r.URL.Path {
		return routing.NotFoundf("asset not found: %v", r.URL.Path)
	}

	http.Redirect(w, r, dest, http.StatusFound)
	return nil
}

type mapConfig struct {
	cache time.Duration
}

type MapOption func(cnf *mapConfig)

func WithCache(cache time.Duration) MapOption {
	return func(cnf *mapConfig) {
		cnf.cache = cache
	}
}

func MapSingleFile(name string, opts ...MapOption) routing.Handler {
	cnf := new(mapConfig)
	for _, opt := range opts {
		opt(cnf)
	}

	return func(w http.ResponseWriter, r *http.Request) error {
		dest, err := Map(name)
		if err != nil {
			return errors.Trace(err)
		}

		if cnf.cache > 0 && !env.IsLocal() {
			w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d", cnf.cache/time.Second))
		} else {
			w.Header().Set("Cache-Control", "private, no-store")
		}
		http.Redirect(w, r, dest, http.StatusFound)
		return nil
	}
}
