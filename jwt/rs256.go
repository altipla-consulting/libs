package jwt

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/altipla-consulting/errors"
	log "github.com/sirupsen/logrus"
	jose "gopkg.in/square/go-jose.v2"
)

type rs256 struct {
	wkurl string

	cancelBackground context.CancelFunc

	mu   sync.RWMutex
	keys jose.JSONWebKeySet
}

func (crypto *rs256) backgroundGetKeys() {
	ctx, cancel := context.WithCancel(context.Background())
	crypto.cancelBackground = cancel

	crypto.mu.Lock()
	for {
		keys, err := crypto.updateKeys(ctx)
		if err != nil {
			slog.Error("Cannot request initial set of signing keys, retrying in 5 seconds...",
				slog.String("error", err.Error()))
			time.Sleep(5 * time.Second)
			continue
		}
		crypto.keys = keys
		break
	}
	crypto.mu.Unlock()

	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		for {
			select {
			case <-ctx.Done():
				return

			case <-ticker.C:
				keys, err := crypto.updateKeys(ctx)
				if err != nil {
					slog.Error("Cannot update set of signing keys, reusing old ones for one more hour...",
						slog.String("error", err.Error()))
					continue
				}

				crypto.mu.Lock()
				crypto.keys = keys
				crypto.mu.Unlock()
			}
		}
	}()
}

func (crypto *rs256) updateKeys(ctx context.Context) (jose.JSONWebKeySet, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, crypto.wkurl, nil)
	if err != nil {
		return jose.JSONWebKeySet{}, errors.Trace(err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return jose.JSONWebKeySet{}, errors.Trace(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return jose.JSONWebKeySet{}, errors.Errorf("unexpected well known url http status %v", resp.Status)
	}

	var keyset jose.JSONWebKeySet
	if err := json.NewDecoder(resp.Body).Decode(&keyset); err != nil {
		return jose.JSONWebKeySet{}, errors.Trace(err)
	}

	var valid jose.JSONWebKeySet
	for _, key := range keyset.Keys {
		// We only accept keys for the RS256 algorithm like the standard or with no
		// algorithm specified like Azure AD.
		if key.Algorithm != string(jose.RS256) && key.Algorithm != "" {
			continue
		}

		// Do not accept private or symmetric keys.
		if !key.IsPublic() {
			continue
		}

		if !key.Valid() {
			log.WithFields(log.Fields{
				"source": crypto.wkurl,
				"kid":    key.KeyID,
			}).Warning("Key obtained from the well known URL is not valid")
			continue
		}

		valid.Keys = append(valid.Keys, key)
	}

	return valid, nil
}

func (crypto *rs256) Signer() (jose.Signer, error) {
	return nil, errors.Errorf("signing with public key is not supported right now")
}

func (crypto *rs256) Key(kid string) interface{} {
	crypto.mu.RLock()
	defer crypto.mu.RUnlock()

	keys := crypto.keys.Key(kid)
	if len(keys) == 0 {
		return nil
	}
	return keys[0]
}

func (crypto *rs256) Close() {
	crypto.cancelBackground()
}
