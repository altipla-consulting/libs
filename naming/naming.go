package naming

import (
	"fmt"
	"strconv"
	"strings"

	"libs.altipla.consulting/errors"
)

func Generate(parts ...interface{}) string {
	if len(parts) == 0 {
		panic("pass parts as argument to generate the name")
	}

	var segments []string
	for _, part := range parts {
		segment := fmt.Sprintf("%v", part)
		if segment != "" {
			segments = append(segments, strings.Trim(segment, "/"))
		}
	}
	return strings.Join(segments, "/")
}

func Read(name string, parts ...interface{}) error {
	if len(parts) == 0 {
		panic("pass parts as argument to read the name")
	}

	segments := strings.Split(name, "/")
	if len(segments) != len(parts) {
		return errors.Errorf("invalid number of segments: %v: %s", len(segments), name)
	}

	for i, part := range parts {
		switch part := part.(type) {
		case string:
			if segments[i] != part {
				return errors.Errorf("mismatch in static segment: %s", part)
			}

		case *string:
			*part = segments[i]

		case *int64:
			n, err := strconv.ParseInt(segments[i], 10, 64)
			if err != nil {
				return errors.Errorf("invalid numeric id: %s", segments[i])
			}
			*part = n

		default:
			panic(fmt.Sprintf("unknown type in name: %#v", part))
		}
	}

	return nil
}
