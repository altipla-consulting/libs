package secrets

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"libs.altipla.consulting/env"
	"libs.altipla.consulting/errors"
)

// Value is a secret accesor that will keep its own value updated in background.
// When read it will return the latest available version of the secret.
type Value struct {
	name  string
	hooks []ChangeHook

	mu      sync.RWMutex
	current []byte
}

// ChangeHook is a function called when a secret changes.
type ChangeHook func(val *Value)

// NewValue creates a new secret accessor.
func NewValue(ctx context.Context, name string) (*Value, error) {
	log.WithField("name", name).Info("Read secret")
	initial, err := readSecret(ctx, name)
	if err != nil {
		return nil, errors.Trace(err)
	}
	val := &Value{
		name:    name,
		current: initial,
	}
	go val.update()
	return val, nil
}

// String gets the current value of the secret as a string.
func (val *Value) String() string {
	val.mu.RLock()
	defer val.mu.RUnlock()
	return string(val.current)
}

// String gets the current value of the secret as a slice of bytes.
func (val *Value) Bytes() []byte {
	val.mu.RLock()
	defer val.mu.RUnlock()
	return append([]byte{}, val.current...)
}

// String gets the current value of the secret and fills a JSON structure.
func (val *Value) JSON(dest interface{}) error {
	val.mu.RLock()
	defer val.mu.RUnlock()
	return errors.Trace(json.Unmarshal(val.current, dest))
}

// OnChange registers a hook to be called when a value change is detected. In the hook
// you can update clients or other actions the application needs.
func (val *Value) OnChange(hook ChangeHook) {
	val.hooks = append(val.hooks, hook)
}

func (val *Value) update() {
	if env.IsLocal() || env.IsCloudRun() {
		return
	}

	for {
		time.Sleep(1 * time.Hour)

		secret, err := readSecret(context.Background(), val.name)
		if err != nil {
			log.WithFields(errors.LogFields(err)).Warning("Cannot update secret. Will retry later.")
			continue
		}
		if val.trySet(secret) {
			for _, hook := range val.hooks {
				hook(val)
			}
		}
	}
}

func (val *Value) trySet(secret []byte) bool {
	val.mu.Lock()
	defer val.mu.Unlock()

	if string(secret) != string(val.current) {
		log.WithField("name", val.name).Info("Read secret")
		val.current = secret
		return true
	}

	return false
}
