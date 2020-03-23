package firestore

import (
	"context"

	"cloud.google.com/go/firestore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"libs.altipla.consulting/errors"
)

type stringKVItem struct {
	Content string `firestore:"content"`
}

type StringKV struct {
	c                *firestore.Client
	collection, name string
}

func (kv *StringKV) Put(ctx context.Context, content string) error {
	item := &stringKVItem{Content: content}
	_, err := kv.c.Collection(kv.collection).Doc(kv.name).Set(ctx, item)
	return errors.Trace(err)
}

func (kv *StringKV) Get(ctx context.Context, content *string) error {
	if content == nil {
		return errors.Errorf("pass a pointer to the result to Get")
	}

	snapshot, err := kv.c.Collection(kv.collection).Doc(kv.name).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return errors.Trace(ErrNoSuchEntity)
		}

		return errors.Trace(err)
	}

	item := new(stringKVItem)
	if err := snapshot.DataTo(item); err != nil {
		return errors.Trace(err)
	}

	*content = item.Content

	return nil
}
