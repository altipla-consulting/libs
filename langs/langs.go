package langs

const (
	CA = "ca"
	DE = "de"
	EN = "en"
	ES = "es"
	EU = "eu"
	FR = "fr"
	IT = "it"
	JA = "ja"
	PT = "pt"
	RU = "ru"
)

var All = []string{
	CA,
	DE,
	EN,
	ES,
	EU,
	FR,
	IT,
	JA,
	PT,
	RU,
}

var native = map[string]string{
	"CA": "Català",
	"DE": "Deutsch",
	"EN": "English",
	"ES": "Español",
	"EU": "Euskera",
	"FR": "Français",
	"IT": "Italiano",
	"JA": "日本語",
	"PT": "Portugues",
	"RU": "русский",
}

func IsValid(lang string) bool {
	for _, l := range All {
		if l == lang {
			return true
		}
	}
	return false
}

func NativeName(lang string) string {
	return native[lang]
}
