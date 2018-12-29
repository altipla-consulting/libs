package funcs

import (
	"bytes"
	"fmt"
	"html"
	"html/template"
	"strings"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
)

type ComponentInstance struct {
	name   string
	params []*ComponentInstanceParam
}

type ComponentInstanceParam struct {
	name  string
	value interface{}
}

func Component(name string) *ComponentInstance {
	return &ComponentInstance{name: name}
}

func EndComponent(c *ComponentInstance) (template.HTML, error) {
	var buf bytes.Buffer

	fmt.Fprintf(&buf, `<div data-vue="%s"`, c.name)
	for _, param := range c.params {
		switch v := param.value.(type) {
		case proto.Message:
			m := jsonpb.Marshaler{EmitDefaults: true}
			b, err := m.MarshalToString(v)
			if err != nil {
				return template.HTML(""), fmt.Errorf("templates: cannot marshal component param %s: %s", c.name, err)
			}
			param.name = fmt.Sprintf("$%v", param.name)
			param.value = html.EscapeString(b)

		case bool, int32, int, float64:
			param.name = fmt.Sprintf("$%v", param.name)

		case string:
			param.value = html.EscapeString(v)

		case int64:
			param.value = fmt.Sprintf("%v", v)

		case []int32:
			var s []string
			for _, n := range v {
				s = append(s, fmt.Sprintf("%d", n))
			}
			param.name = fmt.Sprintf("$%v", param.name)
			param.value = fmt.Sprintf("[%s]", strings.Join(s, ","))
		}

		fmt.Fprintf(&buf, ` data-%v="%v"`, param.name, param.value)
	}
	fmt.Fprintf(&buf, `></div>`)

	return template.HTML(buf.String()), nil
}

func ComponentParam(c *ComponentInstance, name string, value interface{}) string {
	c.params = append(c.params, &ComponentInstanceParam{name, value})
	return ""
}
