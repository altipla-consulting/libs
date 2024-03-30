package schema

import (
	"net/http"
	"reflect"
	"time"

	"github.com/altipla-consulting/errors"
	"github.com/gorilla/schema"

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
		var merr schema.MultiError
		if errors.As(err, &merr) {
			for _, single := range merr {
				var empty schema.EmptyFieldError
				if errors.As(single, &empty) {
					return routing.BadRequestf("required parameter: %v", empty.Key)
				}
			}
		}

		return errors.Trace(err)
	}

	return nil
}
