package sanitize

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFilename(t *testing.T) {
	var tests = []struct {
		filename, result string
	}{
		{"Concierto David Bisbal Straße 03 Feb", "Concierto_David_Bisbal_Strasse_03_Feb"},
		{"Sesión de fotos influencers 10 Mar", "Sesion_de_fotos_influencers_10_Mar"},
		{"Festival jeque µnión deportiva âlmería 09 Dic", "Festival_jeque_nion_deportiva_almeria_09_Dic"},
	}
	for i, test := range tests {
		require.Equal(t, Filename(test.filename), test.result, "iteration %v", i)
	}
}
