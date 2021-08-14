package api

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"libs.altipla.consulting/errors"
)

type Duration time.Duration

func (duration Duration) MarshalJSON() ([]byte, error) {
	t := time.Duration(duration)

	format := []byte(`"`)
	if days := t / (24 * time.Hour); days > 0 {
		format = append(format, fmt.Sprintf("%d.", days)...)
		t -= days * 24 * time.Hour
	}

	hours := t / time.Hour
	t -= hours * time.Hour
	minutes := t / time.Minute
	t -= minutes * time.Minute
	seconds := t / time.Second
	format = append(format, fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)...)

	format = append(format, '"')

	return format, nil
}

func (duration *Duration) UnmarshalJSON(data []byte) error {
	var result time.Duration

	parse, err := strconv.Unquote(string(data))
	if err != nil {
		return errors.Trace(err)
	}

	if parts := strings.Split(parse, "."); len(parts) > 1 {
		n, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			return errors.Trace(err)
		}
		result += time.Duration(n) * 24 * time.Hour
		parse = parts[1]
	}

	parts := strings.Split(parse, ":")
	n, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return errors.Trace(err)
	}
	result += time.Duration(n) * time.Hour
	n, err = strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return errors.Trace(err)
	}
	result += time.Duration(n) * time.Minute
	n, err = strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		return errors.Trace(err)
	}
	result += time.Duration(n) * time.Second

	*duration = Duration(result)
	return nil
}
