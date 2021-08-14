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
		if os.Getenv("FIRESTORE_EMULATOR_HOST") == "" {
			if err := os.Setenv("FIRESTORE_EMULATOR_HOST", "localhost:12000"); err != nil {
				return nil, errors.Trace(err)
			}
		}
	}

	c, err := firestore.NewClient(context.Background(), googleCloudProject)
	if err != nil {
		return nil, errors.Trace(err)
	}

	return &Database{c: c}, nil
}

func (db *Database) StringKV(collection, key string) *StringKV {
	return &StringKV{
		ent: db.Entity(&stringKVEntity{
			collection: collection,
			key:        key,
		}),
		collection: collection,
		key:        key,
	}
}

func (db *Database) ProtoKV(collection, key string) *ProtoKV {
	return &ProtoKV{
		ent: db.Entity(&protoKVEntity{
			collection: collection,
			key:        key,
		}),
		collection: collection,
		key:        key,
	}
}

func (db *Database) Entity(gold Model) *EntityKV {
	return &EntityKV{
		c:    db.c,
		gold: gold,
	}
}
