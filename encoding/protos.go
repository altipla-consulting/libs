package encoding

import (
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// MarshalProtoJSON encode proto message into JSON.
func MarshalProtoJSON(value proto.Message) []byte {
	m := protojson.MarshalOptions{
		EmitUnpopulated: true,
	}
	buf, err := m.Marshal(value)
	if err != nil {
		panic(err)
	}

	return buf
}

// UnmarshalProtoJSON decode JSON into proto message.
func UnmarshalProtoJSON(value []byte, dest proto.Message) {
	if err := protojson.Unmarshal(value, dest); err != nil {
		panic(err)
	}
}
