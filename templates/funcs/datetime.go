package funcs

import (
	"time"

	pbtimestamp "github.com/golang/protobuf/ptypes/timestamp"
	log "github.com/sirupsen/logrus"

	"libs.altipla.consulting/datetime"
	"libs.altipla.consulting/errors"
	"libs.altipla.consulting/langs"
	pbdatetime "libs.altipla.consulting/protos/datetime"
)

var EuropeMadrid *time.Location

func init() {
	var err error
	EuropeMadrid, err = time.LoadLocation("Europe/Madrid")
	if err != nil {
		log.Fatal(err)
	}
}

func commonLayouts(layout string) string {
	switch layout {
	case "time":
		layout = "15:04:05"
	case "datetime":
		layout = "Mon 2 Jan 2006, 15:04:05"
	}
	return layout
}

func DateFormat(t time.Time, lang, layout string) string {
	return datetime.Format(t, lang, layout)
}

func ProtoDateFormat(lang, layout string, t *pbdatetime.Date) string {
	return datetime.Format(datetime.ParseDate(t), lang, layout)
}

func TimestampFormat(layout string, timestamp *pbtimestamp.Timestamp) string {
	return datetime.Format(datetime.ParseTimestamp(timestamp), langs.ES, commonLayouts(layout))
}

func TimeFormat(layout string, t time.Time) string {
	return datetime.Format(t, langs.ES, commonLayouts(layout))
}

func Now() time.Time {
	return time.Now()
}

func Timezone(timezone string, t time.Time) (time.Time, error) {
	if timezone == "Europe/Madrid" {
		return t.In(EuropeMadrid), nil
	}

	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return t, errors.Trace(err)
	}

	return t.In(loc), nil
}

func Madrid(value interface{}) (time.Time, error) {
	var t time.Time
	switch v := value.(type) {
	case time.Time:
		t = v
	case *pbtimestamp.Timestamp:
		t = datetime.ParseTimestamp(v)
	default:
		return t, errors.Errorf("unrecognized time for madrid template function: %#v", value)
	}
	return t.In(EuropeMadrid), nil
}
