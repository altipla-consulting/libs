package validation

import (
	"encoding/json"
	"regexp"
	"strings"
)

var (
	codeRe  = regexp.MustCompile(`^[a-z0-9]+([_.-][a-z0-9]+)*$`)
	emailRe = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
)

func IsValidCode(code string) bool {
	return codeRe.MatchString(code)
}

func IsValidEmail(email string) bool {
	return emailRe.MatchString(email)
}

func IsValidJSON(j string) bool {
	if j == "" {
		return true
	}

	return json.Valid([]byte(j))
}

func NormalizeEmail(email string) string {
	return strings.ToLower(email)
}
