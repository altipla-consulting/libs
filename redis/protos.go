package redis

import (
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func unmarshalProto(raw string, model proto.Message) error {
	if raw[0] == '{' {
		return protojson.Unmarshal([]byte(raw), model)
	}

	return proto.Unmarshal([]byte(raw), model)
}
