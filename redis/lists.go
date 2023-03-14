package redis

import (
	"context"
	"reflect"

	"github.com/altipla-consulting/errors"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type StringsList struct {
	db  *Database
	key string
}

func (list *StringsList) Len(ctx context.Context) (int64, error) {
	return list.db.Cmdable(ctx).LLen(list.key).Result()
}

func (list *StringsList) GetRange(ctx context.Context, start, end int64) ([]string, error) {
	return list.db.Cmdable(ctx).LRange(list.key, start, end).Result()
}

func (list *StringsList) GetAll(ctx context.Context) ([]string, error) {
	return list.GetRange(ctx, 0, -1)
}

func (list *StringsList) Add(ctx context.Context, values []string) error {
	return list.db.Cmdable(ctx).LPush(list.key, values).Err()
}

func (list *StringsList) Remove(ctx context.Context, value string) error {
	return list.db.Cmdable(ctx).LRem(list.key, 1, value).Err()
}

type ProtoList struct {
	db  *Database
	key string
}

func (list *ProtoList) Len(ctx context.Context) (int64, error) {
	return list.db.Cmdable(ctx).LLen(list.key).Result()
}

func (list *ProtoList) GetRange(ctx context.Context, start, end int64, result interface{}) error {
	rt := reflect.TypeOf(result)
	rv := reflect.ValueOf(result)
	msg := reflect.TypeOf((*proto.Message)(nil)).Elem()
	if rt.Kind() != reflect.Ptr || rt.Elem().Kind() != reflect.Slice || !rt.Elem().Elem().Implements(msg) {
		return errors.Errorf("expected a pointer to a slice for the result, received %T", result)
	}

	dest := reflect.MakeSlice(rt.Elem(), 0, 0)

	redisResult, err := list.db.Cmdable(ctx).LRange(list.key, start, end).Result()
	if err != nil {
		return errors.Trace(err)
	}
	for _, item := range redisResult {
		value := reflect.New(rt.Elem().Elem().Elem())
		if err := unmarshalProto(item, value.Interface().(proto.Message)); err != nil {
			return errors.Trace(err)
		}

		dest = reflect.Append(dest, value)
	}

	rv.Elem().Set(dest)

	return nil
}

func (list *ProtoList) GetAll(ctx context.Context, result interface{}) error {
	return list.GetRange(ctx, 0, -1, result)
}

func (list *ProtoList) Add(ctx context.Context, values ...proto.Message) error {
	members := make([]interface{}, len(values))
	for i, value := range values {
		encoded, err := protojson.Marshal(value)
		if err != nil {
			return errors.Trace(err)
		}

		members[i] = string(encoded)
	}

	return list.db.Cmdable(ctx).LPush(list.key, members...).Err()
}

func (list *ProtoList) Remove(ctx context.Context, value proto.Message) error {
	encoded, err := protojson.Marshal(value)
	if err != nil {
		return errors.Trace(err)
	}

	return list.db.Cmdable(ctx).LRem(list.key, 1, string(encoded)).Err()
}
