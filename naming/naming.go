package naming

import (
	"fmt"
	"strconv"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
		return status.Errorf(codes.InvalidArgument, "invalid number of segments: %v", len(parts))
	}

	for i, part := range parts {
		switch part := part.(type) {
		case string:
			if segments[i] != part {
				return status.Errorf(codes.InvalidArgument, "mismatch in static segment: %s", part)
			}

		case *string:
			*part = segments[i]

		case *int64:
			n, err := strconv.ParseInt(segments[i], 10, 64)
			if err != nil {
				return status.Errorf(codes.InvalidArgument, "invalid numeric id: %s", segments[i])
			}
			*part = n

		default:
			panic(fmt.Sprintf("unknown type in name: %#v", part))
		}
	}

	return nil
}
