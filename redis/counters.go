package redis

import (
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

func (c *Counter) Set(value int64) error {
	return c.db.sess.Set(c.key, value, 0).Err()
}

func (c *Counter) Get() (int64, error) {
	result, err := c.db.sess.Get(c.key).Int64()
	if err != nil {
		if err == redis.Nil {
			return 0, nil
		}

		return 0, err
	}

	return result, nil
}

func (c *Counter) Increment() error {
	return c.IncrementBy(1)
}

func (c *Counter) Decrement() error {
	return c.IncrementBy(-1)
}

func (c *Counter) IncrementBy(value int64) error {
	return c.db.sess.IncrBy(c.key, value).Err()
}
