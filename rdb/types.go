package rdb

import (
	"encoding/json"
	"time"

	"libs.altipla.consulting/datetime"
	"libs.altipla.consulting/errors"
)

type Date struct {
	time.Time
}

func NewDate(t time.Time) Date {
	return Date{datetime.TimeToDate(t)}
}

func (value Date) String() string {
	if value.Time.IsZero() {
		return ""
	}
	// Zeros for the nanoseconds is bad on purpose to avoid formatting them.
	return value.Time.In(time.UTC).Format("2006-01-02T15:04:05.0000000Z")
}

func (value Date) MarshalJSON() ([]byte, error) {
	if value.IsZero() {
		return json.Marshal(nil)
	}
	return json.Marshal(value.String())
}

func (value *Date) UnmarshalJSON(data []byte) error {
	if string(data) == "0" {
		value.Time = time.Time{}
		return nil
	}
	return errors.Trace(json.Unmarshal(data, &value.Time))
}

type DateTime struct {
	time.Time
}

func NewDateTime(t time.Time) DateTime {
	return DateTime{t}
}

func (value DateTime) String() string {
	if value.Time.IsZero() {
		return ""
	}
	// Zeros for the nanoseconds is bad on purpose to avoid formatting them.
	return value.Time.In(time.UTC).Format("2006-01-02T15:04:05.0000000Z")
}

func (value DateTime) MarshalJSON() ([]byte, error) {
	if value.IsZero() {
		return json.Marshal(nil)
	}
	return json.Marshal(value.String())
}

func (value *DateTime) UnmarshalJSON(data []byte) error {
	if string(data) == "0" {
		value.Time = time.Time{}
		return nil
	}
	return errors.Trace(json.Unmarshal(data, &value.Time))
}
