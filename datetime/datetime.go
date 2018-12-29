package datetime

import (
	"time"

	"github.com/golang/protobuf/ptypes"
	pbtimestamp "github.com/golang/protobuf/ptypes/timestamp"

	pbdatetime "libs.altipla.consulting/protos/datetime"
)

// DiffDays returns the difference between the two dates in days.
func DiffDays(a, b time.Time) int64 {
	diff := b.Sub(a)

	if diff < 0 {
		return int64((diff - 12*time.Hour) / (24 * time.Hour))
	}

	return int64((diff + 12*time.Hour) / (24 * time.Hour))
}

// TimeToDate returns the remaining time to reach the date.
func TimeToDate(t time.Time) time.Time {
	year, month, day := t.Date()

	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

// SerializeDate converts a Go time.Time to proto considering only the date part.
func SerializeDate(t time.Time) *pbdatetime.Date {
	if t.IsZero() {
		return nil
	}

	year, month, day := t.Date()

	return &pbdatetime.Date{
		Day:   int32(day),
		Month: int32(month),
		Year:  int32(year),
	}
}

// ParseDate converts a proto to Go time.Time.
func ParseDate(date *pbdatetime.Date) time.Time {
	if date == nil {
		return time.Time{}
	}
	return time.Date(int(date.Year), time.Month(date.Month), int(date.Day), 0, 0, 0, 0, time.UTC)
}

// SerializeTimestamp converts a Go time.Time to proto.
func SerializeTimestamp(t time.Time) *pbtimestamp.Timestamp {
	if t.IsZero() {
		return nil
	}

	result, err := ptypes.TimestampProto(t)
	if err != nil {
		panic(err)
	}

	return result
}

// ParseTimestamp converts a proto to Go time.Time.
func ParseTimestamp(timestamp *pbtimestamp.Timestamp) time.Time {
	if timestamp == nil {
		return time.Time{}
	}

	t, err := ptypes.Timestamp(timestamp)
	if err != nil {
		panic(err)
	}

	return t
}
