package redis

import (
	"strconv"

	"github.com/go-redis/redis"
)

type StringsSet struct {
	db  *Database
	key string
}

func (set *StringsSet) Len() (int64, error) {
	return set.db.sess.SCard(set.key).Result()
}

func (set *StringsSet) Members() ([]string, error) {
	result, err := set.db.sess.SMembers(set.key).Result()
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (set *StringsSet) Add(values ...string) error {
	members := make([]interface{}, len(values))
	for i := range values {
		members[i] = values[i]
	}

	return set.db.sess.SAdd(set.key, members...).Err()
}

func (set *StringsSet) Remove(values ...string) error {
	members := make([]interface{}, len(values))
	for i := range values {
		members[i] = values[i]
	}

	return set.db.sess.SRem(set.key, members...).Err()
}

func (set *StringsSet) SortAlpha() ([]string, error) {
	result, err := set.sort(&redis.Sort{Alpha: true})
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (set *StringsSet) sort(sort *redis.Sort) ([]string, error) {
	result, err := set.db.sess.Sort(set.key, sort).Result()
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (set *StringsSet) Contains(value string) (bool, error) {
	return set.db.sess.SIsMember(set.key, value).Result()
}

type Int64Set struct {
	db  *Database
	key string
}

func (set *Int64Set) Len() (int64, error) {
	return set.db.sess.SCard(set.key).Result()
}

func (set *Int64Set) Members() ([]int64, error) {
	rawResult, err := set.db.sess.SMembers(set.key).Result()
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

func (set *Int64Set) Add(values ...int64) error {
	members := make([]interface{}, len(values))
	for i := range values {
		members[i] = strconv.FormatInt(values[i], 10)
	}

	return set.db.sess.SAdd(set.key, members...).Err()
}

func (set *Int64Set) Remove(values ...int64) error {
	members := make([]interface{}, len(values))
	for i := range values {
		members[i] = strconv.FormatInt(values[i], 10)
	}

	return set.db.sess.SRem(set.key, members...).Err()
}

func (set *Int64Set) Contains(value int64) (bool, error) {
	return set.db.sess.SIsMember(set.key, strconv.FormatInt(value, 10)).Result()
}
