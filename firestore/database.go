package firestore

import (
	"context"
	"os"

	"cloud.google.com/go/firestore"
	"libs.altipla.consulting/env"
	"libs.altipla.consulting/errors"
)

type Database struct {
	c *firestore.Client
}

func Open(googleCloudProject string) (*Database, error) {
	if env.IsLocal() {
		googleCloudProject = "local"

		if err := os.Setenv("FIRESTORE_EMULATOR_HOST", "firestore:12000"); err != nil {
			return nil, errors.Trace(err)
		}
	}

	c, err := firestore.NewClient(context.Background(), googleCloudProject)
	if err != nil {
		return nil, errors.Trace(err)
	}

	return &Database{c: c}, nil
}

func (db *Database) StringKV(collection, name string) *StringKV {
	return &StringKV{
		c:          db.c,
		collection: collection,
		name:       name,
	}
}

func (db *Database) ProtoKV(collection, name string) *ProtoKV {
	return &ProtoKV{
		c:          db.c,
		collection: collection,
		name:       name,
	}
}
