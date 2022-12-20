package datetime

import (
	"time"

	"cloud.google.com/go/civil"
	pb "google.golang.org/genproto/googleapis/type/date"
)

func ParseCivilDate(date *pb.Date) civil.Date {
	if date == nil {
		return civil.Date{}
	}

	return civil.Date{
		Year:  int(date.Year),
		Month: time.Month(date.Month),
		Day:   int(date.Day),
	}
}

func SerializeCivilDate(date civil.Date) *pb.Date {
	if date.IsZero() {
		return nil
	}

	return &pb.Date{
		Year:  int32(date.Year),
		Month: int32(date.Month),
		Day:   int32(date.Day),
	}
}
