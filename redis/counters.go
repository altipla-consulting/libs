package redis

import (
	"context"
	"fmt"

	"github.com/go-redis/redis"
)

type Counters struct {
	db  *Database
	key string
}

func (counters *Counters) Key(key string) *Counter {
	return &Counter{
		db:  counters.db,
		key: fmt.Sprintf("%s:%s", counters.key, key),
	}
}

type Counter struct {
	db  *Database
	key string
}

func (c *Counter) Set(ctx context.Context, value int64) error {
	return c.db.Cmdable(ctx).Set(c.key, value, 0).Err()
}

func (c *Counter) Get(ctx context.Context) (int64, error) {
	result, err := c.db.Cmdable(ctx).Get(c.key).Int64()
	if err != nil {
		if err == redis.Nil {
			return 0, nil
		}

		return 0, err
	}

	return result, nil
}

func (c *Counter) Increment(ctx context.Context) (int64, error) {
	return c.IncrementBy(ctx, 1)
}

func (c *Counter) Decrement(ctx context.Context) (int64, error) {
	return c.IncrementBy(ctx, -1)
}

func (c *Counter) IncrementBy(ctx context.Context, value int64) (int64, error) {
	return c.db.Cmdable(ctx).IncrBy(c.key, value).Result()
}
