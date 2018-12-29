package redis

import (
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
)

func unmarshalProto(raw string, model proto.Message) error {
	if raw[0] == '{' {
		return jsonpb.UnmarshalString(raw, model)
	}

	return proto.Unmarshal([]byte(raw), model)
}
