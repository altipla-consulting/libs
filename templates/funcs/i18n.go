package funcs

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/altipla-consulting/content"
	"github.com/altipla-consulting/messageformat"
)

var messages = map[string]map[string]string{}

func init() {
	f, err := os.Open("messages.json")
	if err != nil {
		if os.IsNotExist(err) {
			return
		}

		panic(err)
	}
	defer f.Close()

	if err := json.NewDecoder(f).Decode(&messages); err != nil {
		panic(err)
	}
}

func MsgFormat(lang, format string, params ...interface{}) (string, error) {
	format = Translate(lang, format)

	message, err := messageformat.New(format)
	if err != nil {
		return "", fmt.Errorf("templates: cannot parse messageformat: %s", format)
	}

	res, err := message.Format(lang, params)
	if err != nil {
		return "", fmt.Errorf("templates: cannot run messageformat: %s", format)
	}

	return res, nil
}

func Translate(lang, format string) string {
	if lang == "es" {
		return format
	}

	msg, ok := messages[format]
	if !ok {
		msg = map[string]string{"es": format}
	}

	// En producción se parte correctamente la descripción; pero en desarrollo
	// debemos quitarla partiendo el valor resultante de la cadena de lenguajes.
	return strings.Split(content.LangChain(msg, lang), "//")[0]
}
