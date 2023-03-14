package redis

import (
	"reflect"
	"time"
	"unicode"

	"github.com/altipla-consulting/errors"
)

// Model should be implemented by any model that want to be serializable to redis keys.
type Model interface {
	// IsRedisModel should always return true.
	IsRedisModel() bool
}

// property represents a field of the model mapped to a database key.
type property struct {
	// Name of the key.
	Name string

	// Struct field name.
	Field string

	// Value of the field.
	Value interface{}

	// Pointer to the value of the field.
	ReflectValue reflect.Value
}

func extractModelProps(model Model) ([]*property, error) {
	if !model.IsRedisModel() {
		return nil, errors.Errorf("IsRedisModel should always return true for models")
	}

	v := reflect.ValueOf(model).Elem()
	t := reflect.TypeOf(model).Elem()

	props := []*property{}
	for i := 0; i < t.NumField(); i++ {
		fv := v.Field(i)
		ft := t.Field(i)

		if !startsWithUppercase(ft.Name) {
			continue
		}

		prop := &property{
			Name:         ft.Name,
			Field:        ft.Name,
			Value:        fv.Interface(),
			ReflectValue: fv,
		}

		tag := ft.Tag.Get("redis")
		if tag != "" {
			prop.Name = tag
		}

		if prop.Name == "-" {
			continue
		}

		props = append(props, prop)
	}

	return props, nil
}

func isZero(value interface{}) bool {
	switch v := value.(type) {
	case string:
		return len(v) == 0
	case int32:
		return v == 0
	case int64:
		return v == 0
	case bool:
		return !v
	case time.Time:
		return v.IsZero()
	}

	return false
}

func updatedProps(props []*property, model Model) []*property {
	v := reflect.ValueOf(model).Elem()

	var result []*property
	for _, prop := range props {
		result = append(result, &property{
			Name:         prop.Name,
			Field:        prop.Field,
			Value:        v.FieldByName(prop.Field).Interface(),
			ReflectValue: v.FieldByName(prop.Field),
		})
	}

	return result
}

func startsWithUppercase(s string) bool {
	for _, r := range s {
		return unicode.IsUpper(r)
	}

	return false
}
