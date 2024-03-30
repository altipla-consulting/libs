package secrets

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"
	"time"

	"github.com/altipla-consulting/errors"
	log "github.com/sirupsen/logrus"
)

// Value is a secret accesor that will keep its own value updated in background.
// When read it will return the latest available version of the secret.
type Value struct {
	name   string
	hooks  []ChangeHook
	static bool

	mu         sync.RWMutex
	lastUpdate time.Time
	current    []byte
}

// ChangeHook is a function called when a secret changes.
type ChangeHook func(val *Value)

// NewValue creates a new secret accessor that auto-updates in the background.
func NewValue(ctx context.Context, name string) (*Value, error) {
	log.WithField("name", name).Info("Read secret")
	initial, err := readSecret(ctx, name)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return &Value{
		name:       name,
		lastUpdate: time.Now(),
		current:    initial,
	}, nil
}

// NewStaticValue creates a value from a static string that never updates.
func NewStaticValue(value string) *Value {
	return NewStaticValueBytes([]byte(value))
}

// NewStaticValue creates a value from a static byte slice that never updates.
func NewStaticValueBytes(value []byte) *Value {
	return &Value{
		static:  true,
		current: value,
	}
}

// NewStaticValue creates a value from a static JSON serialized struct that never updates.
func NewStaticValueJSON(src interface{}) (*Value, error) {
	value, err := json.Marshal(src)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return NewStaticValueBytes(value), nil
}

// String gets the current value of the secret as a string.
func (val *Value) String() string {
	return string(val.Bytes())
}

// String gets the current value of the secret as a slice of bytes.
func (val *Value) Bytes() []byte {
	current := val.maybeUpdate()
	if current == nil {
		val.mu.RLock()
		defer val.mu.RUnlock()
		current = val.current
	}
	return append([]byte{}, val.current...)
}

// String gets the current value of the secret and fills a JSON structure.
func (val *Value) JSON(dest interface{}) error {
	return errors.Trace(json.Unmarshal(val.Bytes(), dest))
}

// OnChange registers a hook to be called when a value change is detected. In the hook
// you can update clients or other actions the application needs.
func (val *Value) OnChange(hook ChangeHook) {
	val.hooks = append(val.hooks, hook)
}

func (val *Value) shouldUpdate() bool {
	val.mu.RLock()
	defer val.mu.RUnlock()
	return time.Now().After(val.lastUpdate.Add(1 * time.Hour))
}

func (val *Value) maybeUpdate() []byte {
	if val.static {
		return nil
	}
	if !val.shouldUpdate() {
		return nil
	}

	val.mu.Lock()
	defer val.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	secret, err := readSecret(ctx, val.name)
	if err != nil {
		slog.Warn("Cannot update secret. Will retry later.", slog.String("error", err.Error()))
		return nil
	}

	if string(secret) != string(val.current) {
		log.WithField("name", val.name).Info("Read secret")
		val.current = secret
		for _, hook := range val.hooks {
			hook(val)
		}
	}

	return secret
}
