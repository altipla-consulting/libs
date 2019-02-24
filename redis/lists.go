package redis

import (
	"reflect"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"

	"libs.altipla.consulting/errors"
)

type StringsList struct {
	db  *Database
	key string
}

func (list *StringsList) Len() (int64, error) {
	return list.db.sess.LLen(list.key).Result()
}

func (list *StringsList) GetRange(start, end int64) ([]string, error) {
	return list.db.sess.LRange(list.key, start, end).Result()
}

func (list *StringsList) GetAll() ([]string, error) {
	return list.GetRange(0, -1)
}

func (list *StringsList) Add(values []string) error {
	return list.db.sess.LPush(list.key, values).Err()
}

func (list *StringsList) Remove(value string) error {
	return list.db.sess.LRem(list.key, 1, value).Err()
}

type ProtoList struct {
	db  *Database
	key string
}

func (list *ProtoList) Len() (int64, error) {
	return list.db.sess.LLen(list.key).Result()
}

func (list *ProtoList) GetRange(start, end int64, result interface{}) error {
	rt := reflect.TypeOf(result)
	rv := reflect.ValueOf(result)
	msg := reflect.TypeOf((*proto.Message)(nil)).Elem()
	if rt.Kind() != reflect.Ptr || rt.Elem().Kind() != reflect.Slice || !rt.Elem().Elem().Implements(msg) {
		return errors.Errorf("expected a pointer to a slice for the result, received %T", result)
	}

	dest := reflect.MakeSlice(rt.Elem(), 0, 0)

	redisResult, err := list.db.sess.LRange(list.key, start, end).Result()
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

func (list *ProtoList) GetAll(result interface{}) error {
	return list.GetRange(0, -1, result)
}

func (list *ProtoList) Add(values ...proto.Message) error {
	m := new(jsonpb.Marshaler)
	members := make([]interface{}, len(values))
	for i, value := range values {
		encoded, err := m.MarshalToString(value)
		if err != nil {
			return errors.Trace(err)
		}

		members[i] = encoded
	}

	return list.db.sess.LPush(list.key, members...).Err()
}

func (list *ProtoList) Remove(value proto.Message) error {
	m := new(jsonpb.Marshaler)
	encoded, err := m.MarshalToString(value)
	if err != nil {
		return errors.Trace(err)
	}

	return list.db.sess.LRem(list.key, 1, encoded).Err()
}
