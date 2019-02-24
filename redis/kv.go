package redis

import (
	"strconv"
	"time"

	"github.com/go-redis/redis"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"

	"libs.altipla.consulting/errors"
)

type StringKV struct {
	db  *Database
	key string
}

func (kv *StringKV) Set(value string) error {
	return kv.db.sess.Set(kv.key, value, 0).Err()
}

func (kv *StringKV) SetTTL(value string, ttl time.Duration) error {
	return kv.db.sess.Set(kv.key, value, ttl).Err()
}

func (kv *StringKV) Get() (string, error) {
	result, err := kv.db.sess.Get(kv.key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", ErrNoSuchEntity
		}

		return "", err
	}

	return result, nil
}

func (kv *StringKV) Exists() (bool, error) {
	result, err := kv.db.sess.Exists(kv.key).Result()
	if err != nil {
		return false, err
	}

	return result == 1, nil
}

func (kv *StringKV) Delete() error {
	return kv.db.sess.Del(kv.key).Err()
}

type Int32KV struct {
	db  *Database
	key string
}

func (kv *Int32KV) Set(value int32) error {
	return kv.db.sess.Set(kv.key, value, 0).Err()
}

func (kv *Int32KV) SetTTL(value int32, ttl time.Duration) error {
	return kv.db.sess.Set(kv.key, value, ttl).Err()
}

func (kv *Int32KV) Get() (int32, error) {
	result, err := kv.db.sess.Get(kv.key).Int64()
	if err != nil {
		if err == redis.Nil {
			return 0, ErrNoSuchEntity
		}

		return 0, err
	}

	return int32(result), nil
}

func (kv *Int32KV) Exists() (bool, error) {
	result, err := kv.db.sess.Exists(kv.key).Result()
	if err != nil {
		return false, err
	}

	return result == 1, nil
}

func (kv *Int32KV) Delete() error {
	return kv.db.sess.Del(kv.key).Err()
}

type Int64KV struct {
	db  *Database
	key string
}

func (kv *Int64KV) Set(value int64) error {
	return kv.db.sess.Set(kv.key, value, 0).Err()
}

func (kv *Int64KV) SetTTL(value int64, ttl time.Duration) error {
	return kv.db.sess.Set(kv.key, value, ttl).Err()
}

func (kv *Int64KV) Get() (int64, error) {
	result, err := kv.db.sess.Get(kv.key).Int64()
	if err != nil {
		if err == redis.Nil {
			return 0, ErrNoSuchEntity
		}

		return 0, err
	}

	return int64(result), nil
}

func (kv *Int64KV) Exists() (bool, error) {
	result, err := kv.db.sess.Exists(kv.key).Result()
	if err != nil {
		return false, err
	}

	return result == 1, nil
}

func (kv *Int64KV) Delete() error {
	return kv.db.sess.Del(kv.key).Err()
}

// ProtoKV interacts with a protobuf value key.
type ProtoKV struct {
	db  *Database
	key string
}

// Set changes the value of the key.
func (kv *ProtoKV) Set(value proto.Message) error {
	return kv.SetTTL(value, 0)
}

// SetTTL changes the value of the key with a Time-To-Live.
func (kv *ProtoKV) SetTTL(value proto.Message, ttl time.Duration) error {
	m := new(jsonpb.Marshaler)
	encoded, err := m.MarshalToString(value)
	if err != nil {
		return errors.Trace(err)
	}

	return kv.db.sess.Set(kv.key, encoded, ttl).Err()
}

// Get decodes the value in the provided message.
func (kv *ProtoKV) Get(value proto.Message) error {
	result, err := kv.db.sess.Get(kv.key).Result()
	if err != nil {
		if err == redis.Nil {
			return ErrNoSuchEntity
		}

		return errors.Trace(err)
	}

	return unmarshalProto(result, value)
}

// Exists checks if the key exists previously.
func (kv *ProtoKV) Exists() (bool, error) {
	result, err := kv.db.sess.Exists(kv.key).Result()
	if err != nil {
		return false, err
	}

	return result == 1, nil
}

// Delete the key.
func (kv *ProtoKV) Delete() error {
	return kv.db.sess.Del(kv.key).Err()
}

type BooleanKV struct {
	db  *Database
	key string
}

func (kv *BooleanKV) Set(value bool) error {
	return kv.db.sess.Set(kv.key, value, 0).Err()
}

func (kv *BooleanKV) SetTTL(value bool, ttl time.Duration) error {
	return kv.db.sess.Set(kv.key, value, ttl).Err()
}

func (kv *BooleanKV) Get() (bool, error) {
	result, err := kv.db.sess.Get(kv.key).Result()
	if err != nil {
		if err == redis.Nil {
			return false, ErrNoSuchEntity
		}

		return false, err
	}

	return strconv.ParseBool(result)
}

func (kv *BooleanKV) Exists() (bool, error) {
	result, err := kv.db.sess.Exists(kv.key).Result()
	if err != nil {
		return false, err
	}

	return result == 1, nil
}

func (kv *BooleanKV) Delete() error {
	return kv.db.sess.Del(kv.key).Err()
}

type TimeKV struct {
	db  *Database
	key string
}

func (kv *TimeKV) Set(value time.Time) error {
	rawValue, err := value.MarshalText()
	if err != nil {
		return errors.Trace(err)
	}

	return kv.db.sess.Set(kv.key, string(rawValue), 0).Err()
}

func (kv *TimeKV) SetTTL(value time.Time, ttl time.Duration) error {
	rawValue, err := value.MarshalText()
	if err != nil {
		return errors.Trace(err)
	}

	return kv.db.sess.Set(kv.key, string(rawValue), ttl).Err()
}

func (kv *TimeKV) Get() (time.Time, error) {
	rawResult, err := kv.db.sess.Get(kv.key).Result()
	if err != nil {
		if err == redis.Nil {
			return time.Time{}, ErrNoSuchEntity
		}

		return time.Time{}, err
	}

	result := time.Time{}
	if err := result.UnmarshalText([]byte(rawResult)); err != nil {
		return time.Time{}, err
	}

	return result, nil
}

func (kv *TimeKV) Exists() (bool, error) {
	result, err := kv.db.sess.Exists(kv.key).Result()
	if err != nil {
		return false, err
	}

	return result == 1, nil
}

func (kv *TimeKV) Delete() error {
	return kv.db.sess.Del(kv.key).Err()
}
