package templates

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"math/rand"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ernestoalejo/aeimagesflags"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	pbtimestamp "github.com/golang/protobuf/ptypes/timestamp"
	log "github.com/sirupsen/logrus"
	pbdatetime "libs.altipla.consulting/protos/datetime"

	"libs.altipla.consulting/collections"
	"libs.altipla.consulting/content"
	"libs.altipla.consulting/datetime"
	"libs.altipla.consulting/env"
	"libs.altipla.consulting/errors"
	"libs.altipla.consulting/langs"
	"libs.altipla.consulting/messageformat"
	"libs.altipla.consulting/money"
)

var (
	stdfuncs = template.FuncMap{
		"genrange":  fnGenRange,
		"shuffle":   fnShuffle,
		"limit":     fnLimit,
		"slice":     fnSlice,
		"randitem":  fnRandItem,
		"last":      fnLast,
		"hasstring": collections.HasString,

		"safehtml": fnSafeHTML,
		"safejs":   fnSafeJS,
		"safeurl":  fnSafeURL,
		"safecss":  fnSafeCSS,

		"thumbnail": fnThumbnail,

		"nospaces":  fnNospaces,
		"camelcase": fnCamelcase,
		"nl2br":     fnNl2br,
		"hasprefix": strings.HasPrefix,
		"join":      strings.Join,

		"newvar": fnNewvar,
		"setvar": fnSetvar,
		"getvar": fnGetvar,

		"dict":   fnDict,
		"json":   fnJSON,
		"client": fnClient,

		"even":       fnEven,
		"odd":        fnOdd,
		"trio":       fnTrio,
		"mod":        fnMod,
		"div":        fnDiv,
		"times":      fnTimes,
		"add":        fnAdd,
		"percentage": fnPercentage,
		"bytecount":  fnByteCount,

		"development": fnLocal,
		"local":       fnLocal,
		"version":     fnVersion,

		"rev": fnRev,

		"price": fnPrice,
		"money": fnMoney,

		"now":      fnNow,
		"timezone": fnTimezone,
		"madrid":   fnMadrid,
		"datetime": fnDatetime,

		"include": fnInclude,

		"nativename": langs.NativeName,
		"msgformat":  fnMsgFormat,
		"__":         fnTranslate,
	}

	revManifest = make(map[string]string)
	messages    = make(map[string]map[string]string)

	includeLock  = new(sync.RWMutex)
	includeCache = make(map[string]string)
)

func init() {
	f, err := os.Open("rev-manifest.json")
	if err != nil {
		if os.IsNotExist(err) {
			return
		}
		log.Fatal(err)
	}
	defer f.Close()
	if err := json.NewDecoder(f).Decode(&revManifest); err != nil {
		log.Fatal(err)
	}

	f, err = os.Open("messages.json")
	if err != nil {
		if os.IsNotExist(err) {
			return
		}
		log.Fatal(err)
	}
	defer f.Close()
	if err := json.NewDecoder(f).Decode(&messages); err != nil {
		log.Fatal(err)
	}
}

func fnGenRange(n, start int64) []int64 {
	nums := make([]int64, n)
	for i := range nums {
		nums[i] = int64(i) + start
	}

	return nums
}

func fnShuffle(slice interface{}) interface{} {
	if slice == nil {
		return nil
	}

	value := reflect.ValueOf(slice)
	perm := rand.Perm(value.Len())

	result := reflect.MakeSlice(reflect.TypeOf(slice), value.Len(), value.Len())
	for i, idx := range perm {
		result.Index(i).Set(value.Index(idx))
	}

	return result.Interface()
}

func fnLimit(max int, slice interface{}) interface{} {
	if slice == nil {
		return nil
	}

	value := reflect.ValueOf(slice)
	if value.Len() <= max {
		return value.Interface()
	}

	result := reflect.MakeSlice(reflect.TypeOf(slice), max, max)
	for i := 0; i < max; i++ {
		result.Index(i).Set(value.Index(i))
	}

	return result.Interface()
}

func fnSlice(min, max int, slice interface{}) interface{} {
	if slice == nil {
		return nil
	}

	value := reflect.ValueOf(slice)
	if value.Len() <= min {
		return nil
	}

	if max > value.Len() {
		max = value.Len()
	}

	size := max - min
	result := reflect.MakeSlice(reflect.TypeOf(slice), size, size)
	var current int
	for i := min; i < max; i++ {
		result.Index(current).Set(value.Index(i))
		current++
	}

	return result.Interface()
}

func fnRandItem(slice interface{}) interface{} {
	if slice == nil {
		return nil
	}

	value := reflect.ValueOf(slice)
	if value.Len() == 0 {
		return nil
	}

	return value.Index(rand.Intn(value.Len())).Interface()
}

func fnLast(index int, list interface{}) bool {
	return reflect.ValueOf(list).Len()-1 == index
}

func fnSafeHTML(s ...string) (template.HTML, error) {
	if len(s) == 0 {
		return template.HTML(""), nil
	}
	if len(s) > 1 {
		return template.HTML(""), errors.Errorf("can only sanitize one content at a time")
	}
	return template.HTML(s[0]), nil
}

func fnSafeJS(s ...string) (template.JS, error) {
	if len(s) == 0 {
		return template.JS(""), nil
	}
	if len(s) > 1 {
		return template.JS(""), errors.Errorf("can only sanitize one content at a time")
	}
	return template.JS(s[0]), nil
}

func fnSafeURL(s ...string) (template.URL, error) {
	if len(s) == 0 {
		return template.URL(""), nil
	}
	if len(s) > 1 {
		return template.URL(""), errors.Errorf("can only sanitize one content at a time")
	}
	return template.URL(s[0]), nil
}

func fnSafeCSS(s ...string) (template.CSS, error) {
	if len(s) == 0 {
		return template.CSS(""), nil
	}
	if len(s) > 1 {
		return template.CSS(""), errors.Errorf("can only sanitize one content at a time")
	}
	return template.CSS(s[0]), nil
}

func fnThumbnail(servingURL string, strFlags string) (string, error) {
	if servingURL == "" || strFlags == "" {
		return "", nil
	}

	flags := aeimagesflags.Flags{
		ExpiresDays: 365,
	}
	for _, part := range strings.Split(strFlags, ";") {
		strFlag := strings.Split(part, "=")
		if len(strFlag) != 2 {
			return "", errors.Errorf("all flags should be in the form key=value")
		}

		switch strings.TrimSpace(strFlag[0]) {
		case "width":
			n, err := strconv.ParseUint(strFlag[1], 10, 64)
			if err != nil {
				return "", errors.Wrapf(err, "cannot parse width flag")
			}
			flags.Width = n

		case "height":
			n, err := strconv.ParseUint(strFlag[1], 10, 64)
			if err != nil {
				return "", errors.Wrapf(err, "cannot parse height flag")
			}
			flags.Height = n

		case "square-crop":
			flags.SquareCrop = (strFlag[1] == "true")

		case "smart-square-crop":
			flags.SmartSquareCrop = (strFlag[1] == "true")

		case "original":
			flags.Original = (strFlag[1] == "true")

		case "size":
			n, err := strconv.ParseUint(strFlag[1], 10, 64)
			if err != nil {
				return "", errors.Wrapf(err, "cannot parse size flag")
			}
			flags.Size = n

		default:
			return "", errors.Errorf("unknown image flag: %s", strFlag[0])
		}
	}

	servingURL = strings.Replace(servingURL, "http://", "https://", 1)
	return aeimagesflags.Apply(servingURL, flags), nil
}

func fnNospaces(s string) string {
	return strings.Replace(s, " ", "", -1)
}

func fnCamelcase(s string) string {
	chunks := strings.Split(s, "-")
	for idx, val := range chunks {
		if idx > 0 {
			chunks[idx] = strings.Title(val)
		}
	}
	return strings.Join(chunks, "")
}

func fnNl2br(text string) template.HTML {
	return template.HTML(strings.Replace(template.HTMLEscapeString(text), "\n", "<br>", -1))
}

type RuntimeVar struct {
	Value interface{}
}

func fnNewvar(value interface{}) *RuntimeVar {
	return &RuntimeVar{value}
}

func fnSetvar(v *RuntimeVar, value interface{}) string {
	v.Value = value
	return ""
}

func fnGetvar(v *RuntimeVar) interface{} {
	return v.Value
}

func fnDict(values ...interface{}) (map[string]interface{}, error) {
	if len(values)%2 != 0 {
		return nil, errors.Errorf("dict arguments should be pairs of key,value items")
	}

	dict := make(map[string]interface{}, len(values)/2)
	for i := 0; i < len(values); i += 2 {
		key, ok := values[i].(string)
		if !ok {
			return nil, errors.Errorf("dict keys should be strings")
		}
		dict[key] = values[i+1]
	}
	return dict, nil
}

func fnJSON(obj interface{}) (string, error) {
	msg, ok := obj.(proto.Message)
	if ok {
		m := jsonpb.Marshaler{
			EmitDefaults: true,
		}
		b, err := m.MarshalToString(msg)
		return b, errors.Trace(err)
	}

	b, err := json.Marshal(obj)
	if err != nil {
		return "", errors.Trace(err)
	}
	return string(b), nil
}

func fnClient(obj interface{}) (template.JS, error) {
	str, err := fnJSON(obj)
	if err != nil {
		return template.JS(""), errors.Trace(err)
	}
	return fnSafeJS(str)
}

func fnEven(n int) bool {
	return n%2 == 0
}

func fnOdd(n int) bool {
	return n%2 == 1
}

func fnTrio(n int) bool {
	return (n+1)%3 == 0
}

func fnMod(n, m int) int {
	return n % m
}

func fnDiv(a, b int) int {
	return a / b
}

func fnTimes(a, b int64) int64 {
	return a * b
}

func fnAdd(a, b interface{}) (int64, error) {
	var ai, bi int64

	if n, ok := a.(int64); ok {
		ai = n
	} else if n, ok := a.(int); ok {
		ai = int64(n)
	} else {
		return 0, errors.Errorf("invalid add first argument: %#v", a)
	}

	if n, ok := b.(int64); ok {
		bi = n
	} else if n, ok := b.(int); ok {
		bi = int64(n)
	} else {
		return 0, errors.Errorf("invalid add second argument: %#v", b)
	}

	return ai + bi, nil
}

func fnPercentage(old, current int64) int64 {
	return int64(float64(old-current) / float64(old) * 100.)
}

func fnByteCount(b int64, maxUnit ...string) (string, error) {
	if len(maxUnit) > 1 {
		return "", errors.Errorf("only one max unit argument allowed in bytecount function")
	}
	if len(maxUnit) == 0 {
		maxUnit = []string{"E"}
	}

	const unit = 1024
	const units = "KMGTPE"
	if b < unit {
		return fmt.Sprintf("%d B", b), nil
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		if units[exp] == maxUnit[0][0] {
			break
		}
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(b)/float64(div), units[exp]), nil
}

func fnLocal() bool {
	return env.IsLocal()
}

func fnVersion() string {
	return env.Version()
}

func fnRev(source string) string {
	if m, ok := revManifest[source]; ok {
		return m
	}
	return source
}

func fnPrice(value int32) string {
	return money.FromCents(value).Format(money.FormatConfig{})
}

func fnMoney(currency string, value int32) string {
	return money.FromCents(value).Format(money.Currency(currency))
}

func fnNow() time.Time {
	return time.Now()
}

func readTime(value interface{}) (time.Time, error) {
	switch v := value.(type) {
	case time.Time:
		return v, nil
	case *pbtimestamp.Timestamp:
		return datetime.ParseTimestamp(v), nil
	case *pbdatetime.Date:
		return datetime.ParseDate(v), nil
	default:
		return time.Time{}, errors.Errorf("unrecognized time value: %#v", value)
	}
}

func fnTimezone(timezone string, value interface{}) (time.Time, error) {
	t, err := readTime(value)
	if err != nil {
		return time.Time{}, errors.Trace(err)
	}

	if timezone == "Europe/Madrid" {
		return t.In(datetime.EuropeMadrid()), nil
	}

	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return time.Time{}, errors.Trace(err)
	}
	return t.In(loc), nil
}

func fnMadrid(value interface{}) (time.Time, error) {
	t, err := readTime(value)
	if err != nil {
		return time.Time{}, errors.Trace(err)
	}
	return t.In(datetime.EuropeMadrid()), nil
}

func knownLayouts(layout string) string {
	switch layout {
	case "time":
		layout = "15:04:05"
	case "datetime":
		layout = "Mon 2 Jan 2006, 15:04:05"
	case "rfc3339":
		layout = time.RFC3339
	case "date":
		layout = "2 Jan 2006"
	case "short-time":
		layout = "15:04"
	case "iso8601":
		layout = "2006-01-02"
	}
	return layout
}

func fnDatetime(layout string, args ...interface{}) (string, error) {
	lang := langs.ES
	var value interface{}
	switch len(args) {
	case 0:
		return "", errors.Errorf("value to format required as argument for datetime function")
	case 1:
		value = args[0]
	case 2:
		var ok bool
		lang, ok = args[0].(string)
		if !ok {
			return "", errors.Errorf("unrecognized lang value: %#v", args[0])
		}
		value = args[1]
	default:
		return "", errors.Errorf("only one lang argument allowed in datetime function")
	}

	t, err := readTime(value)
	if err != nil {
		return "", errors.Trace(err)
	}

	return datetime.Format(t, lang, knownLayouts(layout)), nil
}

func fnInclude(path string) (string, error) {
	includeLock.RLock()
	cache, ok := includeCache[path]
	includeLock.RUnlock()

	if !ok || env.IsLocal() {
		content, err := ioutil.ReadFile(path)
		if err != nil {
			return "", errors.Errorf("cannot include file: %v", path)
		}

		includeLock.Lock()
		defer includeLock.Unlock()
		includeCache[path] = string(content)
		cache = includeCache[path]
	}

	return cache, nil
}

func fnMsgFormat(lang, format string, params ...interface{}) (string, error) {
	format = fnTranslate(lang, format)

	message, err := messageformat.New(format)
	if err != nil {
		return "", errors.Wrapf(err, "cannot parse messageformat: %s", format)
	}

	res, err := message.Format(lang, params)
	if err != nil {
		return "", errors.Wrapf(err, "cannot run messageformat: %s", format)
	}

	return res, nil
}

func fnTranslate(lang, source string) string {
	if lang == langs.ES {
		return source
	}

	msg, ok := messages[source]
	if !ok {
		msg = make(map[string]string)
		msg[langs.ES] = source
	}

	// En producción se parte correctamente la descripción; pero en desarrollo
	// debemos quitarla partiendo el valor resultante de la cadena de lenguajes.
	return strings.Split(content.LangChain(msg, lang), "//")[0]
}
