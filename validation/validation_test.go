package validation

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidCodes(t *testing.T) {
	require.True(t, IsValidCode("foo"))
	require.True(t, IsValidCode("foo-bar"))
	require.True(t, IsValidCode("foo.bar"))
	require.True(t, IsValidCode("foo_bar"))
	require.True(t, IsValidCode("foo-bar-baz"))
	require.True(t, IsValidCode("foo34-bar-35"))
}

func TestInvalidCodes(t *testing.T) {
	require.False(t, IsValidCode("foo-"))
	require.False(t, IsValidCode("foo."))
	require.False(t, IsValidCode("foo_"))
	require.False(t, IsValidCode(".foo"))
	require.False(t, IsValidCode("fooBarBaz"))
	require.False(t, IsValidCode("foo-bar-"))
	require.False(t, IsValidCode("con_una_ñ"))
	require.False(t, IsValidCode("foo@bar"))
}

func TestValidEmails(t *testing.T) {
	require.True(t, IsValidEmail("foo@foo.com"))
	require.True(t, IsValidEmail("foo-bar@bar.com"))
	require.True(t, IsValidEmail("foo=bar.foo@baz.es"))
	require.True(t, IsValidEmail("foo_bar#@foo.net"))
	require.True(t, IsValidEmail("foo|bar*baz@foo.xx"))
	require.True(t, IsValidEmail("foo34-bar-35@ac.me"))
	require.True(t, IsValidEmail("foo@bar-gmail.com"))
	require.True(t, IsValidEmail("foo@bar"))
}

func TestInvalidEmails(t *testing.T) {
	require.False(t, IsValidEmail("foo-"))
	require.False(t, IsValidEmail("foo.com"))
	require.False(t, IsValidEmail("fooBarBaz.com@"))
	require.False(t, IsValidEmail("foo@bar_gmail.com"))
}

func TestValidJSONs(t *testing.T) {
	require.True(t, IsValidJSON(""))
	require.True(t, IsValidJSON("{}"))
	require.True(t, IsValidJSON(`{"project": "altipla",	"id": 6}`))
}

func TestInValidJSONs(t *testing.T) {
	require.False(t, IsValidJSON(`"project": "altipla", "id": 6`))
	require.False(t, IsValidJSON(`{"project": "altipla", "id": 6`))
	require.False(t, IsValidJSON(`"project": "altipla",	"id": 6}`))
	require.False(t, IsValidJSON(`{"project": "altipla"	"id": 6}`))
	require.False(t, IsValidJSON(`{"project": "altipla", "id": 6,}`))
}

func TestNormalizeEmails(t *testing.T) {
	require.Equal(t, NormalizeEmail("Developer@ALTIPLA.consulting"), "developer@altipla.consulting")
}
