package cloudrun

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sync"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/idtoken"
	"libs.altipla.consulting/env"
	"libs.altipla.consulting/errors"
)

func NewAuthenticatedHTTPClient() *http.Client {
	return &http.Client{
		Transport: &authTransport{
			ts: make(map[string]oauth2.TokenSource),
		},
	}
}

type authTransport struct {
	mx sync.RWMutex
	ts map[string]oauth2.TokenSource
}

func (tr *authTransport) getTokenSource(audience string) oauth2.TokenSource {
	tr.mx.RLock()
	defer tr.mx.RUnlock()
	return tr.ts[audience]
}

func (tr *authTransport) createTokenSource(ctx context.Context, audience string) (oauth2.TokenSource, error) {
	tr.mx.Lock()
	defer tr.mx.Unlock()

	if tr.ts[audience] != nil {
		return tr.ts[audience], nil
	}

	var err error
	tr.ts[audience], err = NewTokenSource(ctx, audience)
	if err != nil {
		return nil, errors.Trace(err)
	}

	return tr.ts[audience], nil
}

func (tr *authTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	aud := fmt.Sprintf("%s://%s", req.URL.Scheme, req.URL.Host)
	ts := tr.getTokenSource(aud)
	if ts == nil {
		var err error
		ts, err = tr.createTokenSource(req.Context(), aud)
		if err != nil {
			return nil, errors.Trace(err)
		}
	}

	token, err := ts.Token()
	if err != nil {
		return nil, errors.Trace(err)
	}

	if idToken, ok := token.Extra("id_token").(string); ok {
		req.Header.Set("Authorization", "Bearer "+idToken)
	} else {
		return nil, errors.Errorf("cannot read id token from source")
	}

	return http.DefaultTransport.RoundTrip(req)
}

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
