package encoding

import (
    "bytes"

    "github.com/golang/protobuf/jsonpb"
    "github.com/golang/protobuf/proto"
)

// MarshalProtoJSON encode proto message into JSON.
func MarshalProtoJSON(value proto.Message) []byte {
    m := jsonpb.Marshaler{
        EmitDefaults: true,
    }
    var buf bytes.Buffer
    if err := m.Marshal(&buf, value); err != nil {
        panic(err)
    }

    return buf.Bytes()
}

// UnmarshalProtoJSON decode JSON into proto message.
func UnmarshalProtoJSON(value []byte, dest proto.Message) {
    m := jsonpb.Unmarshaler{
        AllowUnknownFields: true,
    }
    if err := m.Unmarshal(bytes.NewReader(value), dest); err != nil {
        panic(err)
    }
}