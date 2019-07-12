package redis

import (
	"context"
	"strconv"

	"github.com/go-redis/redis"
)

type StringsSet struct {
	db  *Database
	key string
}

func (set *StringsSet) Len(ctx context.Context) (int64, error) {
	return set.db.Cmdable(ctx).SCard(set.key).Result()
}

func (set *StringsSet) Members(ctx context.Context) ([]string, error) {
	result, err := set.db.Cmdable(ctx).SMembers(set.key).Result()
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (set *StringsSet) Add(ctx context.Context, values ...string) error {
	members := make([]interface{}, len(values))
	for i := range values {
		members[i] = values[i]
	}

	return set.db.Cmdable(ctx).SAdd(set.key, members...).Err()
}

func (set *StringsSet) Remove(ctx context.Context, values ...string) error {
	members := make([]interface{}, len(values))
	for i := range values {
		members[i] = values[i]
	}

	return set.db.Cmdable(ctx).SRem(set.key, members...).Err()
}

func (set *StringsSet) SortAlpha(ctx context.Context) ([]string, error) {
	result, err := set.sort(ctx, &redis.Sort{Alpha: true})
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (set *StringsSet) sort(ctx context.Context, sort *redis.Sort) ([]string, error) {
	result, err := set.db.Cmdable(ctx).Sort(set.key, sort).Result()
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (set *StringsSet) Contains(ctx context.Context, value string) (bool, error) {
	return set.db.Cmdable(ctx).SIsMember(set.key, value).Result()
}

type Int64Set struct {
	db  *Database
	key string
}

func (set *Int64Set) Len(ctx context.Context) (int64, error) {
	return set.db.Cmdable(ctx).SCard(set.key).Result()
}

func (set *Int64Set) Members(ctx context.Context) ([]int64, error) {
	rawResult, err := set.db.Cmdable(ctx).SMembers(set.key).Result()
	if err != nil {
		return nil, err
	}

	result := []int64{}
	for _, r := range rawResult {
		n, err := strconv.ParseInt(r, 10, 64)
		if err != nil {
			return nil, err
		}

		result = append(result, n)
	}

	return result, nil
}

func (set *Int64Set) Add(ctx context.Context, values ...int64) error {
	members := make([]interface{}, len(values))
	for i := range values {
		members[i] = strconv.FormatInt(values[i], 10)
	}

	return set.db.Cmdable(ctx).SAdd(set.key, members...).Err()
}

func (set *Int64Set) Remove(ctx context.Context, values ...int64) error {
	members := make([]interface{}, len(values))
	for i := range values {
		members[i] = strconv.FormatInt(values[i], 10)
	}

	return set.db.Cmdable(ctx).SRem(set.key, members...).Err()
}

func (set *Int64Set) Contains(ctx context.Context, value int64) (bool, error) {
	return set.db.Cmdable(ctx).SIsMember(set.key, strconv.FormatInt(value, 10)).Result()
}
