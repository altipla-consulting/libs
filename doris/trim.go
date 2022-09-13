package doris

import (
	"strings"

	"google.golang.org/protobuf/reflect/protoreflect"
)

func trimMessage(m protoreflect.Message) protoreflect.Message {
	m.Range(func(fd protoreflect.FieldDescriptor, value protoreflect.Value) bool {
		switch {
		case fd.Kind() == protoreflect.StringKind && fd.IsList():
			for list, i := value.List(), 0; i < list.Len(); i++ {
				list.Set(i, trimString(list.Get(i)))
			}

		case fd.Kind() == protoreflect.StringKind:
			m.Set(fd, trimString(value))

		// We need to trim the keys too, so we build a new map and then replace
		// the existing one once we ranged all the keys.
		case fd.Kind() == protoreflect.MessageKind && fd.IsMap() && fd.MapKey().Kind() == protoreflect.StringKind:
			trimmed := map[string]protoreflect.Value{}
			value.Map().Range(func(mk protoreflect.MapKey, v protoreflect.Value) bool {
				switch fd.MapValue().Kind() {
				case protoreflect.StringKind:
					v = trimString(v)
				case protoreflect.MessageKind:
					trimMessage(v.Message())
				}

				trimmed[trimString(mk.Value()).String()] = v
				value.Map().Clear(mk)
				return true
			})
			for k, sub := range trimmed {
				value.Map().Set(protoreflect.ValueOfString(k).MapKey(), sub)
			}

		case fd.Kind() == protoreflect.MessageKind && fd.IsMap():
			// Map has numeric keys, only process the values.
			value.Map().Range(func(mk protoreflect.MapKey, v protoreflect.Value) bool {
				switch fd.MapValue().Kind() {
				case protoreflect.StringKind:
					v.Map().Set(mk, trimString(v))
				case protoreflect.MessageKind:
					trimMessage(v.Message())
				}

				return true
			})

		case fd.Kind() == protoreflect.MessageKind && fd.IsList():
			for list, i := value.List(), 0; i < list.Len(); i++ {
				list.Set(i, protoreflect.ValueOfMessage(trimMessage(list.Get(i).Message())))
			}

		case fd.Kind() == protoreflect.MessageKind:
			trimMessage(value.Message())
		}

		return true
	})

	return m
}

func trimString(v protoreflect.Value) protoreflect.Value {
	return protoreflect.ValueOfString(strings.TrimSpace(v.String()))
}
