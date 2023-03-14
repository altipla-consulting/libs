package messageformat

import (
	"bytes"
	"strconv"

	"github.com/altipla-consulting/errors"

	"libs.altipla.consulting/langs"
	"libs.altipla.consulting/messageformat/parse"
)

// MessageFormat instance containing a message that can be formatted.
type MessageFormat struct {
	t *parse.Tree
}

// New parses the message using the MessageFormat specification.
func New(message string) (*MessageFormat, error) {
	t, err := parse.Parse(message)
	if err != nil {
		return nil, errors.Trace(err)
	}

	return &MessageFormat{t}, nil
}

// Format the message to replace the params and select the plurals and genders.
func (msg *MessageFormat) Format(lang string, params []interface{}) (string, error) {
	if !langs.IsValid(lang) {
		return "", errors.Errorf("lang is not valid: %s", lang)
	}

	var buf bytes.Buffer
	s := &state{
		t:    msg.t,
		wr:   &buf,
		vars: make(map[string]interface{}),
		lang: langs.Lang(lang),
	}
	for i, param := range params {
		pos := strconv.FormatInt(int64(i), 10)
		s.vars[pos] = param
	}
	if err := s.execute(); err != nil {
		return "", errors.Trace(err)
	}

	return buf.String(), nil
}
