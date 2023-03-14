package firestore

import (
	"context"
	"fmt"

	"github.com/altipla-consulting/errors"
	"google.golang.org/protobuf/proto"
)

type protoKVEntity struct {
	Content []byte `firestore:"content"`

	collection, key string
}

func (kv *protoKVEntity) Collection() string {
	return kv.collection
}

func (kv *protoKVEntity) Key() string {
	return kv.key
}

type ProtoKV struct {
	ent             *EntityKV
	collection, key string
}

func (kv *ProtoKV) Put(ctx context.Context, model proto.Message) error {
	encoded, err := proto.Marshal(model)
	if err != nil {
		return errors.Trace(err)
	}

	item := &protoKVEntity{
		Content:    encoded,
		collection: kv.collection,
		key:        kv.key,
	}

	if err := kv.ent.Put(ctx, item); err != nil {
		return fmt.Errorf("key %s/%s: %w", kv.collection, kv.key, err)
	}

	return nil
}

func (kv *ProtoKV) Delete(ctx context.Context) error {
	item := &protoKVEntity{
		collection: kv.collection,
		key:        kv.key,
	}
	return errors.Trace(kv.ent.Delete(ctx, item))
}

func (kv *ProtoKV) Get(ctx context.Context, model proto.Message) error {
	content, err := kv.GetBytes(ctx)
	if err != nil {
		return errors.Trace(err)
	}
	return errors.Trace(proto.Unmarshal(content, model))
}

func (kv *ProtoKV) GetBytes(ctx context.Context) ([]byte, error) {
	item := &protoKVEntity{
		collection: kv.collection,
		key:        kv.key,
	}
	if err := kv.ent.Get(ctx, item); err != nil {
		return nil, errors.Trace(err)
	}

	return item.Content, nil
}
