package firestore

import (
	"context"

	"cloud.google.com/go/firestore"
	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"libs.altipla.consulting/errors"
)

type protoKVItem struct {
	Content []byte `firestore:"content"`
}

type ProtoKV struct {
	c                *firestore.Client
	collection, name string
}

func (kv *ProtoKV) Put(ctx context.Context, model proto.Message) error {
	encoded, err := proto.Marshal(model)
	if err != nil {
		return errors.Trace(err)
	}

	item := &protoKVItem{Content: encoded}
	_, err = kv.c.Collection(kv.collection).Doc(kv.name).Set(ctx, item)
	return errors.Trace(err)
}

func (kv *ProtoKV) Delete(ctx context.Context) error {
	_, err := kv.c.Collection(kv.collection).Doc(kv.name).Delete(ctx)
	return errors.Trace(err)
}

func (kv *ProtoKV) Get(ctx context.Context, model proto.Message) error {
	content, err := kv.GetBytes(ctx)
	if err != nil {
		return errors.Trace(err)
	}
	return errors.Trace(proto.Unmarshal(content, model))
}

func (kv *ProtoKV) GetBytes(ctx context.Context) ([]byte, error) {
	snapshot, err := kv.c.Collection(kv.collection).Doc(kv.name).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, errors.Wrapf(ErrNoSuchEntity, "key: %v/%v", kv.collection, kv.name)
		}

		return nil, errors.Trace(err)
	}

	item := new(protoKVItem)
	if err := snapshot.DataTo(item); err != nil {
		return nil, errors.Trace(err)
	}

	return item.Content, nil
}
