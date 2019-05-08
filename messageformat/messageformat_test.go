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

		// Simple replacements.
		{"before {0} middle {1} after", langs.ES, []interface{}{"zero", "one"}, "before zero middle one after"},
		{"before {1} middle {0} after", langs.ES, []interface{}{"zero", "one"}, "before one middle zero after"},
		{"before {0} after", langs.ES, []interface{}{3}, "before 3 after"},

		// Simple plurals.
		{"{0, plural, one {1 persona} other {{0} personas}}", langs.ES, []interface{}{1}, "1 persona"},
		{"{0, plural, one {1 persona} other {{0} personas}}", langs.ES, []interface{}{2}, "2 personas"},
		{"{0, plural, one {1 persona} other {{0} personas}}", langs.ES, []interface{}{3}, "3 personas"},

		// French has a special n=0 plural.
		{"{0, plural, one {{0} persona} other {{0} personas}}", langs.FR, []interface{}{0}, "0 persona"},
		{"{0, plural, one {{0} persona} other {{0} personas}}", langs.FR, []interface{}{1}, "1 persona"},
		{"{0, plural, one {{0} persona} other {{0} personas}}", langs.FR, []interface{}{2}, "2 personas"},

		// Specific plurals.
		{"{0, plural, =1 {first}}", langs.ES, []interface{}{1}, "first"},
		{"{0, plural, =1 {first priority} one {second priority}}", langs.ES, []interface{}{1}, "first priority"},
		{"{0, plural, one {second priority} =1 {first priority}}", langs.ES, []interface{}{1}, "first priority"},

		// Plurals priority, last one wins in generic and first specific one wins.
		{"{0, plural, one {second priority} one {first priority}}", langs.ES, []interface{}{1}, "first priority"},
		{"{0, plural, =1 {first priority} =1 {second priority}}", langs.ES, []interface{}{1}, "first priority"},

		// Escape special chars.
		{"escaped '' simple", langs.ES, []interface{}{}, "escaped ' simple"},
		{"escaped '{' open", langs.ES, []interface{}{}, "escaped { open"},
		{"escaped '}' close", langs.ES, []interface{}{}, "escaped } close"},
		{"escaped '{}{}' both", langs.ES, []interface{}{}, "escaped {}{} both"},

		// Plural with recent variable interpolation.
		{"{0, plural, other {# personas}}", langs.ES, []interface{}{5}, "5 personas"},
		{"{0, plural, =3 {# personas}}", langs.ES, []interface{}{3}, "3 personas"},

		// Plurals withing plurals.
		{"{0, plural, one {{1, plural, one {foo}}}}", langs.ES, []interface{}{1, 1}, "foo"},
		{"{0, plural, other {{1, plural, one {foo #}} #}}", langs.ES, []interface{}{2, 1}, "foo 1 2"},
	}
	for _, item := range items {
		mf, err := New(item.message)
		require.NoError(t, err, item.message)

		result, err := mf.Format(item.lang, item.params)
		require.NoError(t, err, item.message)
		require.Equal(t, result, item.expected)
	}
}
