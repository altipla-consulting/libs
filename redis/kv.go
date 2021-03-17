package redis

import (
	"context"
	"strconv"
	"time"

	"github.com/go-redis/redis"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"libs.altipla.consulting/errors"
)

type StringKV struct {
	db  *Database
	key string
}

func (kv *StringKV) Set(ctx context.Context, value string) error {
	return kv.db.Cmdable(ctx).Set(kv.key, value, 0).Err()
}

func (kv *StringKV) SetTTL(ctx context.Context, value string, ttl time.Duration) error {
	return kv.db.Cmdable(ctx).Set(kv.key, value, ttl).Err()
}

func (kv *StringKV) Get(ctx context.Context) (string, error) {
	result, err := kv.db.Cmdable(ctx).Get(kv.key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", ErrNoSuchEntity
		}

		return "", err
	}

	return result, nil
}

func (kv *StringKV) Exists(ctx context.Context) (bool, error) {
	result, err := kv.db.Cmdable(ctx).Exists(kv.key).Result()
	if err != nil {
		return false, err
	}

	return result == 1, nil
}

func (kv *StringKV) Delete(ctx context.Context) error {
	return kv.db.Cmdable(ctx).Del(kv.key).Err()
}

type Int32KV struct {
	db  *Database
	key string
}

func (kv *Int32KV) Set(ctx context.Context, value int32) error {
	return kv.db.Cmdable(ctx).Set(kv.key, value, 0).Err()
}

func (kv *Int32KV) SetTTL(ctx context.Context, value int32, ttl time.Duration) error {
	return kv.db.Cmdable(ctx).Set(kv.key, value, ttl).Err()
}

func (kv *Int32KV) Get(ctx context.Context) (int32, error) {
	result, err := kv.db.Cmdable(ctx).Get(kv.key).Int64()
	if err != nil {
		if err == redis.Nil {
			return 0, ErrNoSuchEntity
		}

		return 0, err
	}

	return int32(result), nil
}

func (kv *Int32KV) Exists(ctx context.Context) (bool, error) {
	result, err := kv.db.Cmdable(ctx).Exists(kv.key).Result()
	if err != nil {
		return false, err
	}

	return result == 1, nil
}

func (kv *Int32KV) Delete(ctx context.Context) error {
	return kv.db.Cmdable(ctx).Del(kv.key).Err()
}

type Int64KV struct {
	db  *Database
	key string
}

func (kv *Int64KV) Set(ctx context.Context, value int64) error {
	return kv.db.Cmdable(ctx).Set(kv.key, value, 0).Err()
}

func (kv *Int64KV) SetTTL(ctx context.Context, value int64, ttl time.Duration) error {
	return kv.db.Cmdable(ctx).Set(kv.key, value, ttl).Err()
}

func (kv *Int64KV) Get(ctx context.Context) (int64, error) {
	result, err := kv.db.Cmdable(ctx).Get(kv.key).Int64()
	if err != nil {
		if err == redis.Nil {
			return 0, ErrNoSuchEntity
		}

		return 0, err
	}

	return int64(result), nil
}

func (kv *Int64KV) Exists(ctx context.Context) (bool, error) {
	result, err := kv.db.Cmdable(ctx).Exists(kv.key).Result()
	if err != nil {
		return false, err
	}

	return result == 1, nil
}

func (kv *Int64KV) Delete(ctx context.Context) error {
	return kv.db.Cmdable(ctx).Del(kv.key).Err()
}

// ProtoKV interacts with a protobuf value key.
type ProtoKV struct {
	db  *Database
	key string
}

// Set changes the value of the key.
func (kv *ProtoKV) Set(ctx context.Context, value proto.Message) error {
	return kv.SetTTL(ctx, value, 0)
}

// SetTTL changes the value of the key with a Time-To-Live.
func (kv *ProtoKV) SetTTL(ctx context.Context, value proto.Message, ttl time.Duration) error {
	encoded, err := protojson.Marshal(value)
	if err != nil {
		return errors.Trace(err)
	}

	return kv.db.Cmdable(ctx).Set(kv.key, encoded, ttl).Err()
}

// Get decodes the value in the provided message.
func (kv *ProtoKV) Get(ctx context.Context, value proto.Message) error {
	result, err := kv.db.Cmdable(ctx).Get(kv.key).Result()
	if err != nil {
		if err == redis.Nil {
			return ErrNoSuchEntity
		}

		return errors.Trace(err)
	}

	return unmarshalProto(result, value)
}

// Exists checks if the key exists previously.
func (kv *ProtoKV) Exists(ctx context.Context) (bool, error) {
	result, err := kv.db.Cmdable(ctx).Exists(kv.key).Result()
	if err != nil {
		return false, err
	}

	return result == 1, nil
}

// Delete the key.
func (kv *ProtoKV) Delete(ctx context.Context) error {
	return kv.db.Cmdable(ctx).Del(kv.key).Err()
}

type BooleanKV struct {
	db  *Database
	key string
}

func (kv *BooleanKV) Set(ctx context.Context, value bool) error {
	return kv.db.Cmdable(ctx).Set(kv.key, value, 0).Err()
}

func (kv *BooleanKV) SetTTL(ctx context.Context, value bool, ttl time.Duration) error {
	return kv.db.Cmdable(ctx).Set(kv.key, value, ttl).Err()
}

func (kv *BooleanKV) Get(ctx context.Context) (bool, error) {
	result, err := kv.db.Cmdable(ctx).Get(kv.key).Result()
	if err != nil {
		if err == redis.Nil {
			return false, ErrNoSuchEntity
		}

		return false, err
	}

	return strconv.ParseBool(result)
}

func (kv *BooleanKV) Exists(ctx context.Context) (bool, error) {
	result, err := kv.db.Cmdable(ctx).Exists(kv.key).Result()
	if err != nil {
		return false, err
	}

	return result == 1, nil
}

func (kv *BooleanKV) Delete(ctx context.Context) error {
	return kv.db.Cmdable(ctx).Del(kv.key).Err()
}

type TimeKV struct {
	db  *Database
	key string
}

func (kv *TimeKV) Set(ctx context.Context, value time.Time) error {
	rawValue, err := value.MarshalText()
	if err != nil {
		return errors.Trace(err)
	}

	return kv.db.Cmdable(ctx).Set(kv.key, string(rawValue), 0).Err()
}

func (kv *TimeKV) SetTTL(ctx context.Context, value time.Time, ttl time.Duration) error {
	rawValue, err := value.MarshalText()
	if err != nil {
		return errors.Trace(err)
	}

	return kv.db.Cmdable(ctx).Set(kv.key, string(rawValue), ttl).Err()
}

func (kv *TimeKV) Get(ctx context.Context) (time.Time, error) {
	rawResult, err := kv.db.Cmdable(ctx).Get(kv.key).Result()
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

func (kv *TimeKV) Exists(ctx context.Context) (bool, error) {
	result, err := kv.db.Cmdable(ctx).Exists(kv.key).Result()
	if err != nil {
		return false, err
	}

	return result == 1, nil
}

func (kv *TimeKV) Delete(ctx context.Context) error {
	return kv.db.Cmdable(ctx).Del(kv.key).Err()
}
