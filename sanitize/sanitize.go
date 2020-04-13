package sanitize

import (
	"bytes"
)

var transliterations = map[rune]string{
	'À': "A",
	'Á': "A",
	'Â': "A",
	'Ã': "A",
	'Ä': "A",
	'Å': "AA",
	'Æ': "AE",
	'Ç': "C",
	'È': "E",
	'É': "E",
	'Ê': "E",
	'Ë': "E",
	'Ì': "I",
	'Í': "I",
	'Î': "I",
	'Ï': "I",
	'Ð': "D",
	'Ł': "L",
	'Ñ': "N",
	'Ò': "O",
	'Ó': "O",
	'Ô': "O",
	'Õ': "O",
	'Ö': "OE",
	'Ø': "OE",
	'Œ': "OE",
	'Ù': "U",
	'Ú': "U",
	'Ü': "UE",
	'Û': "U",
	'Ý': "Y",
	'Þ': "TH",
	'ẞ': "SS",
	'à': "a",
	'á': "a",
	'â': "a",
	'ã': "a",
	'ä': "ae",
	'å': "aa",
	'æ': "ae",
	'ç': "c",
	'è': "e",
	'é': "e",
	'ê': "e",
	'ë': "e",
	'ì': "i",
	'í': "i",
	'î': "i",
	'ï': "i",
	'ð': "d",
	'ł': "l",
	'ñ': "n",
	'ń': "n",
	'ò': "o",
	'ó': "o",
	'ô': "o",
	'õ': "o",
	'ō': "o",
	'ö': "oe",
	'ø': "oe",
	'œ': "oe",
	'ś': "s",
	'ù': "u",
	'ú': "u",
	'û': "u",
	'ū': "u",
	'ü': "ue",
	'ý': "y",
	'ÿ': "y",
	'ż': "z",
	'þ': "th",
	'ß': "ss",
}

func Filename(filename string) string {
	var buf bytes.Buffer
	var underscore bool
	for _, r := range filename {
		if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			_, _ = buf.WriteRune(r)
			underscore = false
		} else if val, ok := transliterations[r]; ok {
			_, _ = buf.WriteString(val)
			underscore = false
		} else if !underscore {
			_, _ = buf.WriteString("_")
			underscore = true
		}
	}
	return buf.String()
}
