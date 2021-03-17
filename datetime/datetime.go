package datetime

import (
	"fmt"
	"sync"
	"time"

	pbtimestamp "google.golang.org/protobuf/types/known/timestamppb"

	pbdatetime "libs.altipla.consulting/protos/datetime"
)

var (
	europeMadrid *time.Location
	locOnce      sync.Once
)

func EuropeMadrid() *time.Location {
	locOnce.Do(func() {
		var err error
		europeMadrid, err = time.LoadLocation("Europe/Madrid")
		if err != nil {
			panic(fmt.Sprintf("cannot load location Europe/Madrid: %v", err))
		}
	})

	return europeMadrid
}

// DiffDays returns the difference between the two dates in days.
func DiffDays(a, b time.Time) int64 {
	diff := b.Sub(a)

	if diff < 0 {
		return int64((diff - 12*time.Hour) / (24 * time.Hour))
	}

	return int64((diff + 12*time.Hour) / (24 * time.Hour))
}

// TimeToDate returns only the date part of the time removing any hour, minutes or seconds.
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

	return pbtimestamp.New(t)
}

// ParseTimestamp converts a proto to Go time.Time.
func ParseTimestamp(timestamp *pbtimestamp.Timestamp) time.Time {
	if timestamp == nil {
		return time.Time{}
	}

	return timestamp.AsTime()
}

// DateIsAfter returns if date "a" is after date "b".
func DateIsAfter(a, b *pbdatetime.Date) bool {
	return !DateIsEqual(a, b) && !DateIsBefore(a, b)
}

// DateIsEqual returns if date "a" is equal to date "b".
func DateIsEqual(a, b *pbdatetime.Date) bool {
	return !DateIsBefore(a, b) && !DateIsBefore(b, a)
}

// DateIsAfter returns if date "a" is before date "b".
func DateIsBefore(a, b *pbdatetime.Date) bool {
	if a == nil {
		return b != nil
	}

	switch {
	case a.Year < b.Year:
		return true

	case a.Year == b.Year && a.Month < b.Month:
		return true

	case a.Year == b.Year && a.Month == b.Month && a.Day < b.Day:
		return true
	}

	return false
}
