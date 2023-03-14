package firestore

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
	"github.com/altipla-consulting/errors"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var Done = iterator.Done

type Query = firestore.Query

type EntityKV struct {
	c    *firestore.Client
	gold Model
}

func (kv *EntityKV) assertSame(model Model) error {
	if model.Collection() != kv.gold.Collection() {
		return errors.Errorf("expected %T and got %T", kv.gold, model)
	}
	return nil
}

func (kv *EntityKV) Put(ctx context.Context, model Model) error {
	if err := kv.assertSame(model); err != nil {
		return errors.Trace(err)
	}

	_, err := kv.c.Collection(model.Collection()).Doc(model.Key()).Set(ctx, model)
	return errors.Trace(err)
}

func (kv *EntityKV) Delete(ctx context.Context, model Model) error {
	_, err := kv.c.Collection(model.Collection()).Doc(model.Key()).Delete(ctx)
	return errors.Trace(err)
}

func (kv *EntityKV) Get(ctx context.Context, model Model) error {
	snapshot, err := kv.c.Collection(model.Collection()).Doc(model.Key()).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return fmt.Errorf("key %v/%v: %w", model.Collection(), model.Key(), ErrNoSuchEntity)
		}

		return errors.Trace(err)
	}

	if err := snapshot.DataTo(model); err != nil {
		return errors.Trace(err)
	}

	return nil
}

func (kv *EntityKV) Query() Query {
	return kv.c.Collection(kv.gold.Collection()).Query
}
