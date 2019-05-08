package messageformat

import (
	"libs.altipla.consulting/langs"
	"libs.altipla.consulting/messageformat/parse"
)

// Extract language rules as needed from:
// http://www.unicode.org/cldr/charts/latest/supplemental/language_plural_rules.html

func matchesPlural(lang string, c *parse.PluralCase, n int64) bool {
	switch lang {
	case langs.ES, langs.EN, langs.EU, langs.IT, langs.DE:
		return matchesPluralCaseN1(c, n)
	case langs.FR:
		return matchesPluralCaseFrench(c, n)
	}

	panic("unsupported message format lang, please add it to plural_cases.go")
}

func matchesPluralCaseN1(c *parse.PluralCase, n int64) bool {
	if n == 1 {
		return c.Category == parse.PluralOne
	}
	return c.Category == parse.PluralOther
}

func matchesPluralCaseFrench(c *parse.PluralCase, n int64) bool {
	if n == 1 || n == 0 {
		return c.Category == parse.PluralOne
	}
	return c.Category == parse.PluralOther
}
