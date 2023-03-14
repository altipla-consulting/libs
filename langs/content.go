package langs

import (
	"encoding/json"

	"github.com/altipla-consulting/errors"
)

// Content represents a translatable string that can store a different value in
// each language.
//
// It can be serialized to JSON to send it to a client application. It can also
// be used with libs.altipla.consulting/rdb.
type Content struct {
	v map[Lang]string
}

// EmptyContent returns an empty content without any values.
func EmptyContent() Content {
	return Content{}
}

// NewContent builds a new content with a single translated value.
func NewContent(lang Lang, value string) Content {
	content := EmptyContent()
	content.Set(lang, value)
	return content
}

// NewContentFromMap builds a new content from a map containing one or multiple values.
func NewContentFromMap(values map[Lang]string) Content {
	return Content{
		v: values,
	}
}

func (content *Content) init() {
	if content.v == nil {
		content.v = make(map[Lang]string)
	}
}

// Set changes the translated value of a language. If empty that language will be
// discarded from the content.
func (content *Content) Set(lang Lang, value string) {
	content.init()
	if value == "" {
		delete(content.v, lang)
	} else {
		content.v[lang] = value
	}
}

// Get returns the translated value for a language or empty if not present.
func (content Content) Get(lang Lang) string {
	if content.v == nil {
		return ""
	}
	return content.v[lang]
}

// IsEmpty returns if the content does not contains any translated value in any language.
func (content Content) IsEmpty() bool {
	return len(content.v) == 0
}

// Clear removes a specific language translated value.
func (content *Content) Clear(lang Lang) {
	content.init()
	delete(content.v, lang)
}

// ClearAll removes all translated values in all languages.
func (content *Content) ClearAll() {
	content.v = nil
}

// Chain helps configuring the chain of fallbacks for a project.
type Chain struct {
	fallbacks []Lang
}

// ChainOption configures the chain when creating a new one.
type ChainOption func(chain *Chain)

// WithFallbacks configures the fallbacks languages of the chain if the requested
// one is not available. Multiple languages can be specified and the order of declaration
// will be the priority.
func WithFallbacks(langs ...Lang) ChainOption {
	return func(chain *Chain) {
		chain.fallbacks = append(chain.fallbacks, langs...)
	}
}

// NewChain initializes a new chain.
func NewChain(opts ...ChainOption) *Chain {
	chain := new(Chain)
	for _, opt := range opts {
		opt(chain)
	}
	return chain
}

// GetChain does the following steps:
// 1. Return the content in the requested lang if available.
// 2. Use the fallback languages if one of them is available. Order is important here.
// 3. Return any lang available randomly to have something.
//
// If the content is empty it returns an empty string.
func (content Content) GetChain(lang Lang, chain *Chain) string {
	if content.IsEmpty() {
		return ""
	}

	value, ok := content.v[lang]
	if ok {
		return value
	}

	for _, l := range chain.fallbacks {
		value, ok := content.v[l]
		if ok {
			return value
		}
	}

	for k := range content.v {
		return content.v[k]
	}

	panic("should not reach here")
}

// MarshalJSON implements the JSON interface.
func (content Content) MarshalJSON() ([]byte, error) {
	v := content.v
	if v == nil {
		v = make(map[Lang]string)
	}
	return json.Marshal(v)
}

// UnmarshalJSON implements the JSON interface.
func (content *Content) UnmarshalJSON(b []byte) error {
	v := make(map[Lang]string)
	if err := json.Unmarshal(b, &v); err != nil {
		return errors.Trace(err)
	}

	content.v = v
	return nil
}
