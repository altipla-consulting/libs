package security

import (
	"context"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/idtoken"

	"libs.altipla.consulting/env"
	"libs.altipla.consulting/errors"
)

func NewTokenSource(ctx context.Context, audience string) (oauth2.TokenSource, error) {
	if audience == "" {
		return nil, errors.Errorf("must supply a non-empty audience")
	}

	switch {
	case env.IsCloudRun() || os.Getenv("GOOGLE_APPLICATION_CREDENTIALS") != "":
		ts, err := idtoken.NewTokenSource(ctx, audience)
		return ts, errors.Trace(err)

	case env.IsLocal():
		ts, err := google.DefaultTokenSource(ctx, audience)
		if err != nil {
			return nil, err
		}
		return oauth2.ReuseTokenSource(nil, ts), nil

	default:
		return nil, errors.Errorf("cannot detect the source to provide authentication")
	}
}
