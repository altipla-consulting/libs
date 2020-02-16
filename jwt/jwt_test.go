package jwt

import (
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
