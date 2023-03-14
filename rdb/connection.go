package rdb

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"

	"github.com/altipla-consulting/errors"
	log "github.com/sirupsen/logrus"

	"libs.altipla.consulting/rdb/api"
)

type connection struct {
	configmux sync.RWMutex
	client    *http.Client
	base      *url.URL

	debug  bool
	dbname string
}

func newConnection(address, dbname string, debug bool) (*connection, error) {
	base, err := url.Parse(address)
	if err != nil {
		return nil, errors.Trace(err)
	}

	conn := &connection{
		base:   base,
		dbname: dbname,
		debug:  debug,
		client: &http.Client{
			Timeout:   60 * time.Second,
			Transport: http.DefaultTransport,
		},
	}
	if conn.debug {
		conn.client.Transport = &debugTransport{conn.client.Transport}
	}

	return conn, nil
}

func newSecureConnection(credentials Credentials, dbname string, debug bool) (*connection, error) {
	conn, err := newConnection(credentials.Address, dbname, debug)
	if err != nil {
		return nil, errors.Trace(err)
	}

	rootCAs := x509.NewCertPool()
	rootCAs.AppendCertsFromPEM([]byte(credentials.CACert))
	cert, err := tls.X509KeyPair([]byte(credentials.Cert), []byte(credentials.Key))
	if err != nil {
		return nil, errors.Trace(err)
	}

	conn.client.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{
			Certificates: []tls.Certificate{cert},
			RootCAs:      rootCAs,
		},
	}
	if conn.debug {
		conn.client.Transport = &debugTransport{conn.client.Transport}
	}

	return conn, nil
}

type debugTransport struct {
	rt http.RoundTripper
}

func (tr *debugTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	b, err := httputil.DumpRequestOut(r, true)
	if err != nil {
		return nil, errors.Trace(err)
	}
	log.Println("==========")
	log.Println(string(b))
	log.Println("==========")

	response, err := tr.rt.RoundTrip(r)
	if err != nil {
		return nil, errors.Trace(err)
	}

	b, err = httputil.DumpResponse(response, true)
	if err != nil {
		return nil, errors.Trace(err)
	}
	log.Println("==========")
	log.Println(string(b))
	log.Println("==========")

	return response, nil
}

func (conn *connection) sendRequest(ctx context.Context, r *http.Request) (*http.Response, error) {
	conn.configmux.RLock()
	defer conn.configmux.RUnlock()

	r = r.WithContext(ctx)
	r.Header.Set("User-Agent", "ravendb-go-client/4.0.0")
	r.Header.Set("Raven-Client-Version", "4.0.0")

	resp, err := conn.client.Do(r)
	if err != nil {
		return nil, errors.Trace(err)
	}

	switch resp.StatusCode {
	case http.StatusNotFound, http.StatusOK, http.StatusCreated, http.StatusNoContent, http.StatusConflict:
		return resp, nil

	default:
		err := NewUnexpectedStatusError(r, resp)
		_ = resp.Body.Close()
		return nil, err
	}
}

func (conn *connection) buildGET(path string, args map[string]interface{}) (*http.Request, error) {
	conn.configmux.RLock()
	defer conn.configmux.RUnlock()

	q := make(url.Values)
	for k, v := range args {
		switch v := v.(type) {
		case string:
			q.Set(k, v)
		case []string:
			for _, item := range v {
				q.Add(k, item)
			}
		default:
			return nil, errors.Errorf("unrecognized get parameter: %T", v)
		}
	}
	u := &url.URL{
		Scheme:   conn.base.Scheme,
		Host:     conn.base.Host,
		Path:     path,
		RawQuery: q.Encode(),
	}

	return http.NewRequest(http.MethodGet, u.String(), nil)
}

func (conn *connection) buildDELETE(path string, args map[string]string) (*http.Request, error) {
	conn.configmux.RLock()
	defer conn.configmux.RUnlock()

	q := make(url.Values)
	for k, v := range args {
		q.Set(k, v)
	}
	u := &url.URL{
		Scheme:   conn.base.Scheme,
		Host:     conn.base.Host,
		Path:     path,
		RawQuery: q.Encode(),
	}

	return http.NewRequest(http.MethodDelete, u.String(), nil)
}

func (conn *connection) buildPUT(path string, args map[string]string, body interface{}) (*http.Request, error) {
	conn.configmux.RLock()
	defer conn.configmux.RUnlock()

	q := make(url.Values)
	for k, v := range args {
		q.Set(k, v)
	}
	u := &url.URL{
		Scheme:   conn.base.Scheme,
		Host:     conn.base.Host,
		Path:     path,
		RawQuery: q.Encode(),
	}

	var reader io.Reader
	if body != nil {
		var buf bytes.Buffer
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			return nil, errors.Trace(err)
		}
		reader = &buf
	}

	r, err := http.NewRequest(http.MethodPut, u.String(), reader)
	if err != nil {
		return nil, errors.Trace(err)
	}

	if body != nil {
		r.Header.Set("Content-Type", "application/json; charset=UTF-8")
	}

	return r, nil
}

func (conn *connection) buildPOST(path string, args map[string]interface{}, body interface{}) (*http.Request, error) {
	conn.configmux.RLock()
	defer conn.configmux.RUnlock()

	q := make(url.Values)
	for k, v := range args {
		switch v := v.(type) {
		case string:
			q.Set(k, v)
		case []string:
			for _, item := range v {
				q.Add(k, item)
			}
		default:
			return nil, errors.Errorf("unrecognized get parameter: %T", v)
		}
	}
	u := &url.URL{
		Scheme:   conn.base.Scheme,
		Host:     conn.base.Host,
		Path:     path,
		RawQuery: q.Encode(),
	}

	var reader io.Reader
	if body != nil {
		var buf bytes.Buffer
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			return nil, errors.Trace(err)
		}
		reader = &buf
	}

	r, err := http.NewRequest(http.MethodPost, u.String(), reader)
	if err != nil {
		return nil, errors.Trace(err)
	}

	if body != nil {
		r.Header.Set("Content-Type", "application/json; charset=UTF-8")
	}

	return r, nil
}

func (conn *connection) buildPATCH(path string, args map[string]string, body interface{}) (*http.Request, error) {
	conn.configmux.RLock()
	defer conn.configmux.RUnlock()

	q := make(url.Values)
	for k, v := range args {
		q.Set(k, v)
	}
	u := &url.URL{
		Scheme:   conn.base.Scheme,
		Host:     conn.base.Host,
		Path:     path,
		RawQuery: q.Encode(),
	}

	var reader io.Reader
	if body != nil {
		var buf bytes.Buffer
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			return nil, errors.Trace(err)
		}
		reader = &buf
	}

	r, err := http.NewRequest(http.MethodPatch, u.String(), reader)
	if err != nil {
		return nil, errors.Trace(err)
	}

	if body != nil {
		r.Header.Set("Content-Type", "application/json; charset=UTF-8")
	}

	return r, nil
}

func (conn *connection) endpoint(segment string) string {
	return "/databases/" + conn.dbname + "/" + segment
}

func (conn *connection) descriptor(ctx context.Context) (*api.Database, error) {
	r, err := conn.buildGET("/admin/databases", map[string]interface{}{"name": conn.dbname})
	if err != nil {
		return nil, errors.Trace(err)
	}
	resp, err := conn.sendRequest(ctx, r)
	if err != nil {
		return nil, errors.Trace(err)
	}
	defer resp.Body.Close()

	switch {
	case resp.StatusCode == http.StatusNotFound && resp.Header.Get("Database-Missing") == conn.dbname:
		return nil, fmt.Errorf("dbname %q: %w", conn.dbname, ErrDatabaseDoesNotExists)
	case resp.StatusCode == http.StatusOK:
		desc := new(api.Database)
		if err := json.NewDecoder(resp.Body).Decode(desc); err != nil {
			return nil, errors.Trace(err)
		}
		return desc, nil
	default:
		return nil, NewUnexpectedStatusError(r, resp)
	}
}

func (conn *connection) createIndexes(ctx context.Context, indexes []*api.Index) error {
	if len(indexes) == 0 {
		return nil
	}

	req := &api.IndexesRequest{
		Indexes: indexes,
	}
	r, err := conn.buildPUT(conn.endpoint("admin/indexes"), nil, req)
	if err != nil {
		return errors.Trace(err)
	}
	resp, err := conn.sendRequest(ctx, r)
	if err != nil {
		return errors.Trace(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return NewUnexpectedStatusError(r, resp)
	}
	return nil
}
