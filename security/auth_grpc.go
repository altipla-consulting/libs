package security

import (
	"context"
	"fmt"
	"net/url"
	"sync"

	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"

	"libs.altipla.consulting/errors"
)

var (
	defaultCredentials = newCallCredentials()
)

func WithDefaultAuthentication() grpc.DialOption {
	return grpc.WithDefaultCallOptions(grpc.PerRPCCredentials(defaultCredentials))
}

type callCredentials struct {
	cachemu *sync.RWMutex
	cache   map[string]oauth2.TokenSource
}

func newCallCredentials() *callCredentials {
	return &callCredentials{
		cachemu: new(sync.RWMutex),
		cache:   make(map[string]oauth2.TokenSource),
	}
}

func (creds *callCredentials) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	u, err := url.Parse(uri[0])
	if err != nil {
		return nil, errors.Trace(err)
	}
	aud := &url.URL{
		Scheme: u.Scheme,
		Host:   u.Host,
	}

	logger := log.WithField("audience", aud.String())
	logger.WithField("uri", uri).Debug("Request metadata authentication")

	if ts := creds.fromCache(aud.String()); ts != nil {
		logger.Debug("Token source cached previously")

		token, err := ts.Token()
		if err != nil {
			return nil, errors.Trace(err)
		}
		return map[string]string{
			"Authorization": fmt.Sprintf("Bearer %s", token.Extra("id_token")),
		}, nil
	}

	ts, err := NewTokenSource(ctx, aud.String())
	if err != nil {
		return nil, errors.Trace(err)
	}

	creds.cachemu.Lock()
	defer creds.cachemu.Unlock()
	creds.cache[aud.String()] = ts

	token, err := ts.Token()
	if err != nil {
		return nil, errors.Trace(err)
	}
	return map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", token.Extra("id_token")),
	}, nil
}

func (creds *callCredentials) fromCache(audience string) oauth2.TokenSource {
	creds.cachemu.RLock()
	defer creds.cachemu.RUnlock()
	return creds.cache[audience]
}

func (creds *callCredentials) RequireTransportSecurity() bool {
	return false
}
