package security

import (
	"context"

	"cloud.google.com/go/compute/metadata"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"gopkg.in/square/go-jose.v2/jwt"
	"libs.altipla.consulting/errors"
)

type cloudRunTokenSource struct {
	audience string
}

func newCloudRunTokenSource(audience string) oauth2.TokenSource {
	return oauth2.ReuseTokenSource(nil, &cloudRunTokenSource{audience})
}

func (ts *cloudRunTokenSource) Token() (*oauth2.Token, error) {
	idToken, err := metadata.Get("/instance/service-accounts/default/identity?audience=" + ts.audience)
	if err != nil {
		return nil, errors.Trace(err)
	}

	signed, err := jwt.ParseSigned(idToken)
	if err != nil {
		return nil, errors.Trace(err)
	}
	var claims jwt.Claims
	if err := signed.UnsafeClaimsWithoutVerification(&claims); err != nil {
		return nil, errors.Trace(err)
	}

	token := &oauth2.Token{
		Expiry: claims.Expiry.Time(),
	}
	token = token.WithExtra(map[string]interface{}{
		"id_token": idToken,
	})

	return token, nil
}

type gcloudTokenSource struct {
	ts oauth2.TokenSource
}

func newGcloudTokenSource(audience string) (oauth2.TokenSource, error) {
	ts, err := google.DefaultTokenSource(context.Background(), audience)
	if err != nil {
		return nil, err
	}
	return oauth2.ReuseTokenSource(nil, &gcloudTokenSource{ts}), nil
}

func (s *gcloudTokenSource) Token() (*oauth2.Token, error) {
	token, err := s.ts.Token()
	return token, errors.Trace(err)
}
