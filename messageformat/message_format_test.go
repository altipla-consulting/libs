package messageformat

import (
	"testing"

	"github.com/stretchr/testify/require"
	"libs.altipla.consulting/langs"
)

type testItem struct {
	message  string
	lang     string
	params   []interface{}
	expected string
}

func TestFormat(t *testing.T) {
	items := []testItem{
		// No replacements.
		{"string without formatting", langs.ES, []interface{}{}, "string without formatting"},

		// Simple replacement and reversed order.
		{"before {0} middle {1} after", langs.ES, []interface{}{"zero", "one"}, "before zero middle one after"},
		{"before {1} middle {0} after", langs.ES, []interface{}{"zero", "one"}, "before one middle zero after"},

		// Different singular-plural cases in Spanish.
		{"{0, plural, one {1 persona} other {{0} personas}}", langs.ES, []interface{}{0}, "0 personas"},
		{"{0, plural, one {1 persona} other {{0} personas}}", langs.ES, []interface{}{1}, "1 persona"},
		{"{0, plural, one {1 persona} other {{0} personas}}", langs.ES, []interface{}{2}, "2 personas"},

		// Exact case
		{"{0, plural, =3 {3 personas} =7 {7 personas}}", langs.ES, []interface{}{3}, "3 personas"},
		{"{0, plural, =3 {3 personas} =7 {7 personas}}", langs.ES, []interface{}{7}, "7 personas"},
	}
	for _, item := range items {
		mf, err := New(item.message)
		require.NoError(t, err)

		result, err := mf.Format(item.lang, item.params)
		require.NoError(t, err)
		require.Equal(t, result, item.expected)
	}
}
