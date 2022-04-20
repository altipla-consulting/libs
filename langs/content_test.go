package langs

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestContentSetGet(t *testing.T) {
	content := EmptyContent()
	content.Set(ES, "value")
	require.Equal(t, content.Get(ES), "value")
}

func TestContentSetGetEmpty(t *testing.T) {
	var content Content
	content.Set(ES, "value")
	require.Equal(t, content.Get(ES), "value")
}

func TestContentIsEmpty(t *testing.T) {
	var content Content
	require.True(t, content.IsEmpty())
}

func TestContentIsEmptyWithEmptyContent(t *testing.T) {
	var content Content
	content.Set(ES, "value")
	require.False(t, content.IsEmpty())
	content.Set(ES, "")
	require.True(t, content.IsEmpty())
}

func TestContentClearSingleLanguage(t *testing.T) {
	content := NewContent(ES, "value")
	require.False(t, content.IsEmpty())
	content.Clear(ES)
	require.True(t, content.IsEmpty())
}

func TestContentClearMultipleLanguages(t *testing.T) {
	var content Content
	content.Set(EN, "value-en")
	content.Set(ES, "value-es")

	content.Clear(ES)

	require.False(t, content.IsEmpty())
	require.Equal(t, content.Get(EN), "value-en")
	require.Empty(t, content.Get(ES))
}

func TestContentClearAll(t *testing.T) {
	var content Content
	content.Set(EN, "value-en")
	content.Set(ES, "value-es")

	content.ClearAll()

	require.True(t, content.IsEmpty())
}

func TestContentGetChain(t *testing.T) {
	tests := []struct {
		content Content
		lang    Lang
		chain   *Chain
		result  string
	}{
		{
			content: EmptyContent(),
			lang:    ES,
			chain:   NewChain(),
			result:  "",
		},
		{
			content: NewContent(ES, "value-es"),
			lang:    ES,
			chain:   NewChain(),
			result:  "value-es",
		},
		{
			content: NewContentFromMap(map[Lang]string{
				ES: "value-es",
				EN: "value-en",
			}),
			lang:   ES,
			chain:  NewChain(),
			result: "value-es",
		},
		{
			content: NewContentFromMap(map[Lang]string{
				ES: "value-es",
				EN: "value-en",
			}),
			lang:   FR,
			chain:  NewChain(WithFallbacks(EN, ES)),
			result: "value-en",
		},
		{
			content: NewContentFromMap(map[Lang]string{
				ES: "value-es",
				IT: "value-it",
			}),
			lang:   FR,
			chain:  NewChain(WithFallbacks(EN, ES)),
			result: "value-es",
		},
		{
			content: NewContentFromMap(map[Lang]string{
				IT: "value-it",
			}),
			lang:   FR,
			chain:  NewChain(WithFallbacks(EN, ES)),
			result: "value-it",
		},
	}
	for _, test := range tests {
		require.Equal(t, test.content.GetChain(test.lang, test.chain), test.result)
	}
}

func TestContentMarshalJSON(t *testing.T) {
	content := NewContentFromMap(map[Lang]string{
		ES: "value-es",
		IT: "value-it",
	})
	b, err := json.Marshal(content)
	require.NoError(t, err)
	require.JSONEq(t, string(b), `{"es": "value-es", "it": "value-it"}`)
}

func TestContentMarshalJSONEmpty(t *testing.T) {
	content := EmptyContent()
	b, err := json.Marshal(content)
	require.NoError(t, err)
	require.JSONEq(t, string(b), `{}`)
}

func TestContentUnmarshalJSON(t *testing.T) {
	var content Content
	require.NoError(t, json.Unmarshal([]byte(`{"es": "value-es", "it": "value-it"}`), &content))

	require.Equal(t, content.Get(ES), "value-es")
	require.Equal(t, content.Get(IT), "value-it")
}

func TestContentUnmarshalJSONEmpty(t *testing.T) {
	var content Content
	require.NoError(t, json.Unmarshal([]byte(`{}`), &content))

	require.True(t, content.IsEmpty())
}

func TestContentUnmarshalJSONNil(t *testing.T) {
	var content Content
	require.NoError(t, json.Unmarshal([]byte(`null`), &content))

	require.True(t, content.IsEmpty())
}
