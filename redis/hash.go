package redis

import (
	"context"
	"reflect"
	"slices"
	"strconv"
	"time"

	"github.com/altipla-consulting/errors"
)

type Hash struct {
	db    *Database
	name  string
	props []*property
}

type MaskOpt []string

func Mask(cols ...string) []string {
	return cols
}

func (hash *Hash) Get(ctx context.Context, key string, instance Model, masks ...MaskOpt) error {
	var included []string
	for _, mask := range masks {
		included = append(included, mask...)
	}

	modelProps := updatedProps(hash.props, instance)

	var names []string
	var filteredProps []*property
	for _, prop := range modelProps {
		if len(included) > 0 && !slices.Contains(included, prop.Name) {
			continue
		}

		names = append(names, prop.Name)
		filteredProps = append(filteredProps, prop)
	}

	fields, err := hash.db.Cmdable(ctx).HMGet(hash.name+":"+key, names...).Result()
	if err != nil {
		return errors.Trace(err)
	}
	var nilFields int
	for i, field := range fields {
		prop := filteredProps[i]
		if field == nil {
			nilFields++
			continue
		}

		switch prop.Value.(type) {
		case int64:
			n, err := strconv.ParseInt(field.(string), 10, 64)
			if err != nil {
				return errors.Trace(err)
			}
			prop.ReflectValue.Set(reflect.ValueOf(n))

		case string:
			prop.ReflectValue.Set(reflect.ValueOf(field.(string)))

		case time.Time:
			var t time.Time
			if err := t.UnmarshalText([]byte(field.(string))); err != nil {
				return errors.Trace(err)
			}
			prop.ReflectValue.Set(reflect.ValueOf(t))
		}
	}
	if nilFields == len(filteredProps) {
		return ErrNoSuchEntity
	}

	return nil
}

func (hash *Hash) Put(ctx context.Context, key string, instance Model, masks ...MaskOpt) error {
	var included []string
	for _, mask := range masks {
		included = append(included, mask...)
	}

	modelProps := updatedProps(hash.props, instance)

	fields := map[string]interface{}{}
	for _, prop := range modelProps {
		if len(included) > 0 && !slices.Contains(included, prop.Name) {
			continue
		}

		var store string
		switch v := prop.Value.(type) {
		case int64:
			store = strconv.FormatInt(v, 10)
		case string:
			store = v
		case time.Time:
			t, err := v.MarshalText()
			if err != nil {
				return errors.Trace(err)
			}
			store = string(t)
		}

		fields[prop.Name] = store
	}

	if err := hash.db.Cmdable(ctx).HMSet(hash.name+":"+key, fields).Err(); err != nil {
		return errors.Trace(err)
	}

	return nil
}

// Delete inmediately removes the key from the hash.
func (hash *Hash) Delete(ctx context.Context, key string) error {
	if err := hash.db.Cmdable(ctx).Del(hash.name + ":" + key).Err(); err != nil {
		return errors.Trace(err)
	}
	return nil
}

// ExpireAt sets the expiration of a key of this hash.
func (hash *Hash) ExpireAt(ctx context.Context, key string, t time.Time) error {
	if err := hash.db.Cmdable(ctx).ExpireAt(hash.name+":"+key, t).Err(); err != nil {
		return errors.Trace(err)
	}
	return nil
}
