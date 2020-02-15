package database

import (
	"database/sql/driver"

	"libs.altipla.consulting/errors"
)

// NullableString can be used in a model column as a normal string, but if the
// database column is NULL an empty string will be returned. If you want to
// differentiate between those two cases use sql.NullString directly instead.
type NullableString string

// Scan implements the Scanner interface.
func (ns *NullableString) Scan(value interface{}) error {
	if value == nil {
		*ns = ""
		return nil
	}

	switch s := value.(type) {
	case string:
		*ns = NullableString(s)
	case []byte:
		*ns = NullableString(s)
	default:
		return errors.Errorf("type %T is not a string that NullableString can scan: %#v", value, value)
	}

	return nil
}

// Value implements the driver Valuer interface.
func (ns NullableString) Value() (driver.Value, error) {
	return string(ns), nil
}
