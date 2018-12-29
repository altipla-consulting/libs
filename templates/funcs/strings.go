package funcs

import (
	"html/template"
	"strings"
)

func NoSpaces(s string) string {
	return strings.Replace(s, " ", "", -1)
}

func CamelCase(s string) string {
	chunks := strings.Split(s, "-")
	for idx, val := range chunks {
		if idx > 0 {
			chunks[idx] = strings.Title(val)
		}
	}

	return strings.Join(chunks, "")
}

func Nl2Br(text string) template.HTML {
	return template.HTML(strings.Replace(template.HTMLEscapeString(text), "\n", "<br>", -1))
}

func HasPrefix(s, prefix string) bool {
	return strings.HasPrefix(s, prefix)
}

func Join(chain []string, sep string) string {
	return strings.Join(chain, sep)
}
