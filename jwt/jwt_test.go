package jwt

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"

	"libs.altipla.consulting/clock"
	"libs.altipla.consulting/errors"
)

func TestSign(t *testing.T) {
	g := NewHS256("foobarbaz")
	claims := Claims{
		Issuer:   "https://tests.com/",
		Subject:  "foo",
		Audience: "bar",
		Expiry:   time.Date(2019, time.October, 5, 4, 3, 2, 0, time.UTC),
		IssuedAt: time.Date(2019, time.September, 5, 4, 3, 2, 0, time.UTC),
	}
	token, err := g.Sign(claims)
	require.NoError(t, err)

	require.Equal(t, token, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOlsiYmFyIl0sImV4cCI6MTU3MDI0ODE4MiwiaWF0IjoxNTY3NjU2MTgyLCJpc3MiOiJodHRwczovL3Rlc3RzLmNvbS8iLCJzdWIiOiJmb28ifQ.m_ONMeFLHFLG1cR-J2E2C08rg0dRfeZzSsUsQkcaeAM")
}

func TestExtract(t *testing.T) {
	g := NewHS256("foobarbaz")
	g.clock = clock.NewStatic(time.Date(2019, time.October, 1, 4, 3, 2, 0, time.UTC))

	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOlsiYmFyIl0sImV4cCI6MTU3MDI0ODE4MiwiaWF0IjoxNTY3NjU2MTgyLCJpc3MiOiJodHRwczovL3Rlc3RzLmNvbS8iLCJzdWIiOiJmb28ifQ.m_ONMeFLHFLG1cR-J2E2C08rg0dRfeZzSsUsQkcaeAM"
	expected := Expected{
		Issuer:   "https://tests.com/",
		Audience: "bar",
	}
	claims := Claims{}
	require.NoError(t, g.Extract(token, expected, &claims))

	require.Equal(t, claims.Issuer, "https://tests.com/")
	require.Equal(t, claims.Subject, "foo")
	require.Equal(t, claims.Audience, "bar")
	require.WithinDuration(t, claims.Expiry, time.Date(2019, time.October, 5, 4, 3, 2, 0, time.UTC), 1*time.Second)
	require.WithinDuration(t, claims.IssuedAt, time.Date(2019, time.September, 5, 4, 3, 2, 0, time.UTC), 1*time.Second)
}

func TestExtractInvalidParts(t *testing.T) {
	g := NewHS256("foobarbaz")
	g.clock = clock.NewStatic(time.Date(2019, time.October, 1, 4, 3, 2, 0, time.UTC))

	token := "invalid token"
	expected := Expected{
		Issuer:   "https://tests.com/",
		Audience: "bar",
	}
	claims := Claims{}
	log.Println(errors.Details(g.Extract(token, expected, &claims)))
	require.EqualError(t, g.Extract(token, expected, &claims), "jwt: invalid token: invalid parts")
}

func TestExtractInvalidSignature(t *testing.T) {
	g := NewHS256("foobarbaz")
	g.clock = clock.NewStatic(time.Date(2019, time.October, 1, 4, 3, 2, 0, time.UTC))

	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOlsiYmFyIl0sImV4cCI6MTU3MDI0ODE4MiwiaWF0IjoxNTY3NjU2MTgyLCJpc3MiOiJodHRwczovL3Rlc3RzLmNvbS8iLCJzdWIiOiJmb28ifQ.m_ONMeFLHFLG1cR-"
	expected := Expected{
		Issuer:   "https://tests.com/",
		Audience: "bar",
	}
	claims := Claims{}
	require.EqualError(t, g.Extract(token, expected, &claims), "jwt: invalid token: invalid signature")
}

func TestExtractWithSpaces(t *testing.T) {
	g := NewHS256("foobarbaz")
	g.clock = clock.NewStatic(time.Date(2019, time.October, 1, 4, 3, 2, 0, time.UTC))

	token := "  eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOlsiYmFyIl0sImV4cCI6MTU3MDI0ODE4MiwiaWF0IjoxNTY3NjU2MTgyLCJpc3MiOiJodHRwczovL3Rlc3RzLmNvbS8iLCJzdWIiOiJmb28ifQ.m_ONMeFLHFLG1cR-J2E2C08rg0dRfeZzSsUsQkcaeAM  "
	expected := Expected{
		Issuer:   "https://tests.com/",
		Audience: "bar",
	}
	claims := Claims{}
	require.NoError(t, g.Extract(token, expected, &claims))
}

func initJWKServer(t *testing.T) *httptest.Server {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `
			{
				"keys": [
					{"use":"sig","kty":"RSA","kid":"Gb6ymg9WaeIMFJpdXpoVcHOpNlO79bdUmPWh3zn0WpQ=","alg":"RS256","n":"qkrF1_rOZ-AOqCHvKA9BZBDqfJbxANq1QtyCzLagDzq2JcUJ4Z1D5g-5qDdYEhKaojYQBGTCO4dyWlLUSCKGtUWbIgkhNjg2zm7XlG6criPNjRicj1gJ6qhPK0a7hH0bc54C4vLNpYjbdMZ0mzqYPv1iy8pjB0iHkG2nuJ2s_yVt5hxOg8DHuIJGJpdTNWzi7HX1cLauES6GQRhj3myIBy5k82WezHZ9zC_VZLolpRImdzEFW2ckxHxJ1vgPsSfx7fBdZLxtyX6JKPOaHWgO9v_4kgUwSfYRQuLiX3devQBP-ewJvShqeyNEoyGf7tU5SqaBUeCpjzjqTc7w0iOhYrMGwms5gDniTWHh5f_Oh11NMNTezW6RolKRylmo_CEKFv-cBFTkNV-4k-x-mx3pGUrVoCvKGF0xrcI3irLndCbmvRDvPFv2yM6Ekt1t_q58BU4vJg639V1bjVSJDwZq8g7ORvo_6BMP5_egfTapTv7A1s0ZY0DfoZ4xxRVLGvy-5QTTUtsDydhQ6LxyQe7LZVvf2i9LCNeyM1iPXzPWKR8tzLaRVGa9bKhNNVrjGr1NmHwwm319IcZbIC_w-4oBu9ETZOW4rs4c75zSvLyV1J9EtjQL1u4OMKHM09r1Ck99QRkFW7Piq85DfFy6VERVEChehLunwD90QR2Gq-s-Zx8","e":"AQAB"},
					{"use":"sig","kty":"RSA","kid":"MJvEs8dR6SeB9Y_NNmi3jCEJBTWHDRZOSSp5dvkTFpc=","alg":"RS256","n":"xrjbQ1qF5ZP13HCZ9BGsKe1gE9ryteDwdEX0SKpiCaVi6_F3J0T-UlXC77RbmyGVckbwhSyDPLx3BiqKpVLlSLLm9VVTnkccGWpvo-Ws8YXFULVo9U8MHRaN3KElUMhY_xDi8lJoMRADdmVzVrFCUfZbEgGmYc5QY2dRVM47kBYry00MiJvP2NLxPTDSwj1QyAGsNf4iXy9hxsozcF-LFJb6fgDWzI1CVYOUmmy5O-aeFdH5avg57ntjYz_kLV91xGJtPD-eJ0YX2gZnNtTBpKX2vYb8xE3JugV-LtCn0IilpklZ0zxVZCg7MYarHeg7lKIZmWCkEecX8P7gor6S52pDxSI97Y14yMkZlavvhxHortU0cgbdaQchAMm_LZeeedfUjOrT_JQVZydx7Vpx9nWBPx_MBPCYLVx2724KNz2eICeOvfjW3DZlNOpEJQM4_ZPIJ8qMZGTWnt8mKTw2y-KBHffLKYELteT_GvkaUYCdb6xGfii8EkHH_HsYoVhQflmTZk-SY8fLQ40L4zMFmTHEczqS8uSPWoR4V2GkLa2KfiJl0CnOd7bQbrzQvvJksNhJBMWSeAvJQcvfmQlx6gsZACBygM8sUxhMtQCnuRLcYhYnhMJRm5XO1KUsAQ_nNBRGmnDyd9Hu3bmWS-rVBp2C_dXB_Hpfuwj8Os1DTRc","e":"AQAB"}
				]
			}
		`)
	}))

	return server
}

func TestExtractRSAMultipleKeys(t *testing.T) {
	s := initJWKServer(t)
	defer s.Close()

	g := NewRS256FromWellKnown(s.URL)
	defer g.Close()
	g.clock = clock.NewStatic(time.Date(2019, time.October, 1, 4, 3, 2, 0, time.UTC))

	token := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCIsImtpZCI6Ik1KdkVzOGRSNlNlQjlZX05ObWkzakNFSkJUV0hEUlpPU1NwNWR2a1RGcGM9In0.eyJhdWQiOlsiYmFyIl0sImV4cCI6MTU3MDI0ODE4MiwiaWF0IjoxNTY3NjU2MTgyLCJpc3MiOiJodHRwczovL3Rlc3RzLmNvbS8iLCJzdWIiOiJmb28ifQ.ExY6Zio0nYfVf8_ILcGK7asqTN3IOQZU5JvBe1IwOZ-ljSe3-osKE4QoW7VOLF-vkphuv_Ef3f-4OQhh8eQpEjVsaj8T55HODU5kdnsVMrow3i4q7GJ3aEMC8CiiMRWcEAoYOV6JBz44me4BaGUhC7At30QHotc4L4qeMJkDLSkoOo9_IYIW6Zd5lsHXmKebRHQv5cjFHUYUhBFtGXSNA10VsK0MdjX9ftDJaW1j6qoXCua8TLSU96_fCLggy4GwzLeOa4X7liYwzyY9A24hvyMSNxRyWqiQHn-V_--iUaFdOwvjDqFpvt7cguAlVCgQd_7WGwsT9TT_efaxI1Xp41TGlQI0FXOYlTOZWMTFUJYaqoF-GqzJh2i11n0MYiSUVD26uyZfAw5u90HP0ZGVXwDtuouLo6rj7mSYUe_KoaCuuGBvfvjXEchWUwSO50INIW1uCarVmiKUVXL-TdN25BdoxApWb5mdsu71iLp5fLTLllFym2kUQP-xP6UOVDJkOJWIDi4HRQITCkHA9suEmSIHZSAcBZOOc5HnE8vjuAscsj4ennPabw46IC-bpzFqHnSF-7vHgSq41vUQH03GC40QG_oF052fkG_Hurzknb3sejwjRer1EJho4UuNlBveclGPHI-XNszTTRVwB8dNHPPzHh_MOmh9EN1vtFMJhO8"
	expected := Expected{
		Issuer:   "https://tests.com/",
		Audience: "bar",
	}
	claims := Claims{}
	require.NoError(t, g.Extract(token, expected, &claims))

	require.Equal(t, claims.Issuer, "https://tests.com/")
	require.Equal(t, claims.Subject, "foo")
	require.Equal(t, claims.Audience, "bar")
	require.WithinDuration(t, claims.Expiry, time.Date(2019, time.October, 5, 4, 3, 2, 0, time.UTC), 1*time.Second)
	require.WithinDuration(t, claims.IssuedAt, time.Date(2019, time.September, 5, 4, 3, 2, 0, time.UTC), 1*time.Second)
}
