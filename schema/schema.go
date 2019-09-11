package schema

import (
	"net/http"
	"reflect"
	"time"

	"github.com/gorilla/schema"

	"libs.altipla.consulting/errors"
	"libs.altipla.consulting/routing"
)

var (
	decoder = schema.NewDecoder()

	invalidValue = reflect.Value{}
)

func init() {
	decoder.RegisterConverter(reflect.TypeOf(time.Time{}), func(s string) reflect.Value {
		t, err := time.Parse(s, time.RFC3339)
		if err == nil {
			if t.IsZero() {
				return invalidValue
			}

			return reflect.ValueOf(t)
		}

		return invalidValue
	})
}

// Load parses the form in the request and loads the incoming data into the struct
// pointer provided in dst.
func Load(r *http.Request, dst interface{}) error {
	if err := r.ParseForm(); err != nil {
		return errors.Trace(err)
	}

	if err := decoder.Decode(dst, r.Form); err != nil {
		multi, ok := errors.Cause(err).(schema.MultiError)
		if ok {
			for _, single := range multi {
				empty, ok := errors.Cause(single).(schema.EmptyFieldError)
				if ok {
					return routing.BadRequestf("required parameter: %v", empty.Key)
				}
			}
		}

		return errors.Trace(err)
	}

	return nil
}
