package content

import (
	"database/sql/driver"
	"encoding/json"

	"libs.altipla.consulting/errors"
)

type Translated map[string]string

// LangChain runs the lang chain over the contents of the container.
func (content Translated) LangChain(lang string) string {
	return LangChain(content, lang)
}

func (content Translated) Value() (driver.Value, error) {
	if content == nil {
		return "{}", nil
	}

	serialized, err := json.Marshal(content)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot serialize value")
	}

	return serialized, nil
}

func (content *Translated) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.Errorf("cannot scan type into bytes: %T", value)
	}

	if err := json.Unmarshal(b, content); err != nil {
		return errors.Wrapf(err, "cannot scan value")
	}

	if *content == nil {
		*content = Translated{}
	}

	return nil
}

// LangChain returns a value following a prearranged chain of preference.
//
// The chain gives preference to the requested lang, then english and finally spanish
// if no other translation is available.
func LangChain(translations map[string]string, lang string) string {
	if translations[lang] != "" {
		return translations[lang]
	}
	if translations["en"] != "" {
		return translations["en"]
	}
	return translations["es"]
}
