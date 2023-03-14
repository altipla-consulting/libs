package secrets

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"cloud.google.com/go/compute/metadata"
	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"github.com/altipla-consulting/errors"
	log "github.com/sirupsen/logrus"
	pb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"

	"libs.altipla.consulting/env"
)

var (
	cacheLock = new(sync.Mutex)
	cache     = make(map[string][]byte)
	client    *secretmanager.Client
)

// ReadBytes reads a secret as a slice of bytes. Multiple reads of the same secret
// will return a cached version.
func ReadBytes(ctx context.Context, name string) ([]byte, error) {
	log.WithField("name", name).Info("Read secret")
	return readSecret(ctx, name)
}

func readSecret(ctx context.Context, name string) ([]byte, error) {
	cacheLock.Lock()
	defer cacheLock.Unlock()

	if v, ok := cache[name]; ok {
		return v, nil
	}

	if env.IsLocal() {
		local, err := readLocalSecrets()
		if err != nil {
			return nil, errors.Trace(err)
		}
		v, ok := local[name]
		if !ok {
			return nil, errors.Errorf("missing local secret: %s", name)
		}
		cache[name] = []byte(v)
		return []byte(v), nil
	}

	googleProject, err := metadata.ProjectID()
	if err != nil {
		return nil, errors.Trace(err)
	}
	if client == nil {
		client, err = secretmanager.NewClient(ctx)
		if err != nil {
			return nil, errors.Trace(err)
		}
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req := &pb.AccessSecretVersionRequest{
		Name: fmt.Sprintf("projects/%s/secrets/%s/versions/latest", googleProject, name),
	}
	reply, err := client.AccessSecretVersion(ctx, req)
	if err != nil {
		return nil, errors.Trace(err)
	}

	cache[name] = reply.Payload.Data

	return reply.Payload.Data, nil
}

// Read a secret and return its value. Multiple reads of the same secret will
// return a cached version.
func Read(ctx context.Context, name string) (string, error) {
	value, err := ReadBytes(ctx, name)
	if err != nil {
		return "", errors.Trace(err)
	}
	return string(value), nil
}

// ReadJSON reads a secrets and fills a JSON structure with it. Multiple reads
// of the same secret will return a cached version.
func ReadJSON(ctx context.Context, name string, dest interface{}) error {
	value, err := ReadBytes(ctx, name)
	if err != nil {
		return errors.Trace(err)
	}
	return errors.Trace(json.Unmarshal(value, dest))
}
