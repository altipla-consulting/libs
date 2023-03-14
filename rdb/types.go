package rdb

import (
	"encoding/json"
	"time"

	"github.com/altipla-consulting/errors"
	"google.golang.org/protobuf/encoding/protojson"

	"libs.altipla.consulting/datetime"
	pb "libs.altipla.consulting/protos/datetime"
	"libs.altipla.consulting/rdb/api"
)

type Date struct {
	time.Time
}

func NewDate(t time.Time) Date {
	return Date{datetime.TimeToDate(t)}
}

func ZeroDate() Date {
	return Date{}
}

func (value Date) String() string {
	if value.Time.IsZero() {
		return ""
	}
	return datetime.TimeToDate(value.Time).In(time.UTC).Format(api.DateTimeFormat)
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
	if err := json.Unmarshal(data, &value.Time); err != nil {
		m := new(pb.Date)
		if err := protojson.Unmarshal(data, m); err != nil {
			return errors.Trace(err)
		}
		value.Time = datetime.ParseDate(m)
	} else {
		value.Time = datetime.TimeToDate(value.Time)
	}
	return nil
}

type DateTime struct {
	time.Time
}

func NewDateTime(t time.Time) DateTime {
	return DateTime{t}
}

func ZeroDateTime() DateTime {
	return DateTime{}
}

func (value DateTime) String() string {
	if value.Time.IsZero() {
		return ""
	}
	return value.Time.In(time.UTC).Format(api.DateTimeFormat)
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
