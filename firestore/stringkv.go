package firestore

import (
	"context"

	"libs.altipla.consulting/errors"
)

type stringKVEntity struct {
	Content string `firestore:"content"`

	collection, key string
}

func (kv *stringKVEntity) Collection() string {
	return kv.collection
}

func (kv *stringKVEntity) Key() string {
	return kv.key
}

type StringKV struct {
	ent             *EntityKV
	collection, key string
}

func (kv *StringKV) Put(ctx context.Context, content string) error {
	item := &stringKVEntity{
		Content:    content,
		collection: kv.collection,
		key:        kv.key,
	}
	return errors.Trace(kv.ent.Put(ctx, item))
}

func (kv *StringKV) Get(ctx context.Context, content *string) error {
	if content == nil {
		return errors.Errorf("pass a pointer to the result to Get")
	}

	item := &stringKVEntity{
		collection: kv.collection,
		key:        kv.key,
	}
	if err := kv.ent.Get(ctx, item); err != nil {
		return errors.Trace(err)
	}

	*content = item.Content
	return nil
}
