package langs

type Lang string

func (lang Lang) String() string {
	return string(lang)
}

const (
	CA = Lang("ca")
	DE = Lang("de")
	EN = Lang("en")
	ES = Lang("es")
	EU = Lang("eu")
	FR = Lang("fr")
	IT = Lang("it")
	JA = Lang("ja")
	PT = Lang("pt")
	RU = Lang("ru")
)

var All = []Lang{
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

var native = map[Lang]string{
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
		if string(l) == lang {
			return true
		}
	}
	return false
}

func NativeName(lang Lang) string {
	return native[lang]
}
