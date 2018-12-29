package messageformat

// Extract language rules as needed from:
// http://www.unicode.org/cldr/charts/latest/supplemental/language_plural_rules.html

func getPluralCase(lang string, n int64) pluralCaseType {
	switch lang {
	case "es":
		return getPluralCaseN1(n)

	case "en":
		return getPluralCaseN1(n)

	case "it":
		return getPluralCaseN1(n)

	case "de":
		return getPluralCaseN1(n)

	case "fr":
		return getPluralCaseFrench(n)
	}

	panic("unsupported message format lang, please add it to plural_cases.go")
}

func getPluralCaseN1(n int64) pluralCaseType {
	if n == 1 {
		return pluralCaseOne
	}

	return pluralCaseOther
}

func getPluralCaseFrench(n int64) pluralCaseType {
	if n == 1 || n == 0 {
		return pluralCaseOne
	}

	return pluralCaseOther
}
