package slack

import (
	"strings"
)

// EscapeMessage removes any formatting characters from the string to avoid
// unusual strings with third party data.
//
// It is really not an "escape" function but there is no other option because Slack
// doesn't perform any escaping of the input.
func EscapeMessage(msg string) string {
	s := strings.NewReplacer("_", "", "*", "")
	return s.Replace(msg)
}
