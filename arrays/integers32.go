package arrays

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"libs.altipla.consulting/errors"
)

type Integers32 []int32

func (slice Integers32) Value() (driver.Value, error) {
	if slice == nil {
		return "[]", nil
	}

	serialized, err := json.Marshal(slice)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot serialize value")
	}

	return serialized, nil
}

func (slice *Integers32) Scan(value interface{}) error {
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

func SearchIntegers32(column string) string {
	return fmt.Sprintf("JSON_CONTAINS(%s, CAST(? AS CHAR), '$')", column)
}
