package arrays

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"libs.altipla.consulting/errors"
)

type Strings []string

func (slice Strings) Value() (driver.Value, error) {
	if slice == nil {
		return "[]", nil
	}

	serialized, err := json.Marshal(slice)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot serialize value")
	}

	return serialized, nil
}

func (slice *Strings) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	b, ok := value.([]byte)
	if !ok {
		return errors.Errorf("cannot scan type into bytes: %T", value)
	}

	if err := json.Unmarshal(b, slice); err != nil {
		return errors.Wrapf(err, "cannot scan value")
	}

	return nil
}

func SearchStrings(column string) string {
	return fmt.Sprintf("JSON_CONTAINS(%s, JSON_QUOTE(?), '$')", column)
}
