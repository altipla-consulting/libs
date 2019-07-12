package redis

import (
	"context"
	"reflect"

	"github.com/go-redis/redis"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"

	"libs.altipla.consulting/errors"
)

type ProtoHash struct {
	db  *Database
	key string
}

func (hash *ProtoHash) PrepareInsert() *ProtoHashInsert {
	return &ProtoHashInsert{
		hash:   hash,
		fields: make(map[string]interface{}),
	}
}

// GetMulti fetchs a list of keys from the hash. Result should be a slice of proto.Message
// that will be filled with the results in the same order as the keys.
func (hash *ProtoHash) GetMulti(ctx context.Context, keys []string, result interface{}) error {
	if len(keys) == 0 {
		return nil
	}

	rt := reflect.TypeOf(result)
	rv := reflect.ValueOf(result)
	msg := reflect.TypeOf((*proto.Message)(nil)).Elem()
	if rt.Kind() != reflect.Ptr || rt.Elem().Kind() != reflect.Slice || !rt.Elem().Elem().Implements(msg) {
		return errors.Errorf("expected a pointer to a slice for the result, received %T", result)
	}

	dest := reflect.MakeSlice(rt.Elem(), 0, 0)

	var merr MultiError

	redisResult, err := hash.db.Cmdable(ctx).HMGet(hash.key, keys...).Result()
	if err != nil {
		return errors.Trace(err)
	}
	for _, item := range redisResult {
		var model reflect.Value
		if item == nil {
			model = reflect.Zero(rt.Elem().Elem())
			merr = append(merr, ErrNoSuchEntity)
		} else {
			model = reflect.New(rt.Elem().Elem().Elem())
			if err := unmarshalProto(item.(string), model.Interface().(proto.Message)); err != nil {
				return errors.Trace(err)
			}
			merr = append(merr, nil)
		}

		dest = reflect.Append(dest, model)
	}

	rv.Elem().Set(dest)

	if merr.HasError() {
		return merr
	}
	return nil
}

func (hash *ProtoHash) Get(ctx context.Context, key string, model proto.Message) error {
	redisResult, err := hash.db.Cmdable(ctx).HGet(hash.key, key).Result()
	if err != nil {
		if err == redis.Nil {
			return ErrNoSuchEntity
		}

		return errors.Trace(err)
	}

	return unmarshalProto(redisResult, model)
}

func (hash *ProtoHash) Delete(ctx context.Context, key string) error {
	return hash.db.Cmdable(ctx).HDel(hash.key, key).Err()
}

type ProtoHashInsert struct {
	hash   *ProtoHash
	fields map[string]interface{}
}

func (insert *ProtoHashInsert) Set(ctx context.Context, key string, value proto.Message) error {
	m := new(jsonpb.Marshaler)
	encoded, err := m.MarshalToString(value)
	if err != nil {
		return errors.Trace(err)
	}

	insert.fields[key] = encoded

	return nil
}

func (insert *ProtoHashInsert) Commit(ctx context.Context) error {
	return insert.hash.db.Cmdable(ctx).HMSet(insert.hash.key, insert.fields).Err()
}
