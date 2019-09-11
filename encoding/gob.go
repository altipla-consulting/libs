package encoding

import (
	"bytes"
	"encoding/binary"

	"libs.altipla.consulting/errors"
)

// GobInt32 can be used as a pointer inside a struct. When serialized with encoding/gob
// it will differentiate between a nil pointer or a zero, which is not possible with
// the stdlib.
//
// It helps us for example when transfering XML decoded data to a queue, where we need
// to encode the value but keep the information of which fields are present and which
// ones are really zero value.
type GobInt32 int32

// NewGobInt32 builds a new instance with the specified value. It is a comodity
// to build new numbers with a constant without an intermediary variable.
func NewGobInt32(v int32) *GobInt32 {
	x := GobInt32(v)
	return &x
}

// GobEncode implements the encoding/gob interface.
func (v *GobInt32) GobEncode() ([]byte, error) {
	var buf bytes.Buffer

	if v == nil {
		var present bool
		if err := binary.Write(&buf, binary.LittleEndian, &present); err != nil {
			return nil, errors.Trace(err)
		}
	} else {
		present := true
		if err := binary.Write(&buf, binary.LittleEndian, &present); err != nil {
			return nil, errors.Trace(err)
		}
		if err := binary.Write(&buf, binary.LittleEndian, v); err != nil {
			return nil, errors.Trace(err)
		}
	}

	return buf.Bytes(), nil
}

// GobDecode implements the encoding/gob interface.
func (v *GobInt32) GobDecode(data []byte) error {
	buf := bytes.NewReader(data)

	var present bool
	if err := binary.Read(buf, binary.LittleEndian, &present); err != nil {
		return errors.Trace(err)
	}

	if present {
		*v = 0
		if err := binary.Read(buf, binary.LittleEndian, v); err != nil {
			return errors.Trace(err)
		}
	}

	return nil
}
