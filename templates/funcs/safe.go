package funcs

import (
	"fmt"
	"html/template"
)

func SafeHTML(s ...string) (template.HTML, error) {
	if len(s) == 0 {
		return template.HTML(""), nil
	}

	if len(s) > 1 {
		return template.HTML(""), fmt.Errorf("templates: can only sanitize one content at a time")
	}

	return template.HTML(s[0]), nil
}

func SafeJavascript(s ...string) (template.JS, error) {
	if len(s) == 0 {
		return template.JS(""), nil
	}

	if len(s) > 1 {
		return template.JS(""), fmt.Errorf("templates: can only sanitize one content at a time")
	}

	return template.JS(s[0]), nil
}

func SafeURL(s ...string) (template.URL, error) {
	if len(s) == 0 {
		return template.URL(""), nil
	}

	if len(s) > 1 {
		return template.URL(""), fmt.Errorf("templates: can only sanitize one content at a time")
	}

	return template.URL(s[0]), nil
}

func SafeCSS(s ...string) (template.CSS, error) {
	if len(s) == 0 {
		return template.CSS(""), nil
	}

	if len(s) > 1 {
		return template.CSS(""), fmt.Errorf("templates: can only sanitize one content at a time")
	}

	return template.CSS(s[0]), nil
}
