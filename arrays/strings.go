package arrays

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type Strings []string

func (slice Strings) Value() (driver.Value, error) {
	if slice == nil {
		return "[]", nil
	}

	serialized, err := json.Marshal(slice)
	if err != nil {
		return nil, fmt.Errorf("arrays/strings: cannot serialize value: %s", err)
	}

	return serialized, nil
}

func (slice *Strings) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	b, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("arrays/strings: cannot scan type into bytes: %T", value)
	}

	if err := json.Unmarshal(b, slice); err != nil {
		return fmt.Errorf("arrays/strings: cannot scan value: %s", err)
	}

	return nil
}

func SearchStrings(column string) string {
	return fmt.Sprintf("JSON_CONTAINS(%s, JSON_QUOTE(?), '$')", column)
}
