package funcs

import (
	"encoding/json"
	"fmt"
	"html/template"
	"math/rand"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
)

func Dict(values ...interface{}) (map[string]interface{}, error) {
	if len(values)%2 != 0 {
		return nil, fmt.Errorf("templates: dict arguments should be pairs of key,value items")
	}

	dict := make(map[string]interface{}, len(values)/2)
	for i := 0; i < len(values); i += 2 {
		key, ok := values[i].(string)
		if !ok {
			return nil, fmt.Errorf("templates: dict keys should be strings")
		}

		dict[key] = values[i+1]
	}

	return dict, nil
}

func RandID(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, rand.Int())
}

func JSON(obj interface{}) (string, error) {
	msg, ok := obj.(proto.Message)
	if ok {
		m := jsonpb.Marshaler{
			EmitDefaults: true,
		}
		b, err := m.MarshalToString(msg)
		return b, err
	}

	b, err := json.Marshal(obj)
	if err != nil {
		return "", fmt.Errorf("templates: cannot marshal json")
	}

	return string(b), nil
}

func Vue(obj interface{}) (template.JS, error) {
	str, err := JSON(obj)
	if err != nil {
		return template.JS(""), fmt.Errorf("templates: cannot marshal vue: %s", err)
	}

	return SafeJavascript(str)
}
