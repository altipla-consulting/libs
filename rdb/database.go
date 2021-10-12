package rdb

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"libs.altipla.consulting/env"
	"libs.altipla.consulting/errors"
	"libs.altipla.consulting/rdb/api"
)

type Database struct {
	conn *connection

	strongConsistency bool
	debug             bool
	localCreate       bool
	testingMode       bool

	serverCertificatePEM, clientPrivateKeyPEM, clientCertificatePEM string
}

type OpenOption func(db *Database)

// WithStrongConsistency forces every query to wait for the index to be ready
// and with no pending indexing operations.
//
// This is probably only needed in tests and other rapidly changing environments,
// but not a normal production application.
func WithStrongConsistency() OpenOption {
	return func(db *Database) {
		db.strongConsistency = true
	}
}

func WithDebug() OpenOption {
	if !env.IsLocal() {
		panic("should not enable debug mode in production")
	}

	return func(db *Database) {
		db.debug = true
	}
}

// WithLocalCreate creates automatically the database if it doesn't exists. It can
// only be used in a local environment as it applies a pretty large hit to performance.
func WithLocalCreate() OpenOption {
	if !env.IsLocal() {
		panic("should not enable local create in production")
	}

	return func(db *Database) {
		db.localCreate = true
	}
}

// WithTestingMode enables flags and behaviours suited for unit tests. It implies
// WithLocalCreate() and WithStrongConsistency() and also creates a new database for
// each package tests.
//
// We don't use the argument, but we require it to avoid errors when calling this
// function outside tests.
func WithTestingMode(t *testing.T) OpenOption {
	return func(db *Database) {
		db.testingMode = true
		db.localCreate = true
		db.strongConsistency = true
	}
}

// WithSecurity enables authentication with HTTPS certificates, both server side
// and client side.
func WithSecurity(serverCertificatePEM, clientPrivateKeyPEM, clientCertificatePEM string) OpenOption {
	return func(db *Database) {
		if serverCertificatePEM == "" || clientPrivateKeyPEM == "" || clientCertificatePEM == "" {
			panic("do not enable security without certificates")
		}
		db.serverCertificatePEM = serverCertificatePEM
		db.clientPrivateKeyPEM = clientPrivateKeyPEM
		db.clientCertificatePEM = clientCertificatePEM
	}
}

func Open(address, dbname string, opts ...OpenOption) (*Database, error) {
	db := new(Database)
	for _, opt := range opts {
		opt(db)
	}

	// In testing mode create an independent database for each package tests
	dbnameOriginal := dbname
	if db.testingMode {
		dbname += "_" + filepath.Base(filepath.Dir(os.Args[0])) + "_" + strings.Replace(filepath.Base(os.Args[0]), ".", "_", -1)
	}

	var err error
	db.conn, err = newConnection(address, dbname, db.debug)
	if err != nil {
		return nil, errors.Trace(err)
	}

	if db.serverCertificatePEM != "" {
		if err := db.conn.enableSecurity(db.serverCertificatePEM, db.clientPrivateKeyPEM, db.clientCertificatePEM); err != nil {
			return nil, errors.Trace(err)
		}
	}

	if db.localCreate {
		if exists, err := db.Exists(context.Background()); err != nil {
			return nil, errors.Trace(err)
		} else if !exists {
			if err := db.Create(context.Background(), 1); err != nil {
				return nil, errors.Trace(err)
			}
		}
	}

	// Clone indexes from the original database in testing mode
	if db.testingMode {
		desc, err := db.conn.descriptor(context.Background())
		if err != nil {
			return nil, errors.Trace(err)
		}

		connOriginal, err := newConnection(address, dbnameOriginal, db.debug)
		if err != nil {
			return nil, errors.Trace(err)
		}
		descOriginal, err := connOriginal.descriptor(context.Background())
		if err != nil {
			return nil, errors.Trace(err)
		}
		var clone []*api.Index
		for _, index := range descOriginal.Indexes {
			if _, ok := desc.Indexes[index.Name]; ok {
				continue
			}
			clone = append(clone, index)
		}
		if err := db.conn.createIndexes(context.Background(), clone); err != nil {
			return nil, errors.Trace(err)
		}
	}

	return db, nil
}

// EnableSecurity enables the HTTPS certificates configured when opening a connection.
// It helps transitioning from an insecure database to a secure one with feature flags.
func (db *Database) EnableSecurity() error {
	return errors.Trace(db.conn.enableSecurity(db.serverCertificatePEM, db.clientPrivateKeyPEM, db.clientCertificatePEM))
}

// DisableSecurity disables the HTTPS certificates configured in the connection.
// It helps transitioning from an insecure database to a secure one with feature flags.
func (db *Database) DisableSecurity() error {
	return errors.Trace(db.conn.disableSecurity())
}

func (db *Database) Exists(ctx context.Context) (bool, error) {
	if _, err := db.conn.descriptor(ctx); err != nil {
		if errors.Is(err, ErrDatabaseDoesNotExists) {
			return false, nil
		}
		return false, errors.Trace(err)
	}

	return true, nil
}

func (db *Database) Create(ctx context.Context, replicationFactor int64) error {
	// TODO(ernesto): Probablemente la implementación de este método debería estar en connection
	params := map[string]string{
		"name":              db.conn.dbname,
		"replicationFactor": strconv.FormatInt(replicationFactor, 10),
	}
	r, err := db.conn.buildPUT("/admin/databases", params, &api.Database{DatabaseName: db.conn.dbname})
	if err != nil {
		return errors.Trace(err)
	}
	resp, err := db.conn.sendRequest(ctx, r)
	if err != nil {
		return errors.Trace(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return NewUnexpectedStatusError(r, resp)
	}
	return nil
}

func (db *Database) NewSession(ctx context.Context) (context.Context, *Session) {
	sess := &Session{
		conn: db.conn,
	}
	ctx = context.WithValue(ctx, keySession, sess)
	return ctx, sess
}

func checkFields(v reflect.Value) {
	switch v.Kind() {
	case reflect.Ptr:
		checkFields(v.Elem())

	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			field := v.Field(i)
			if !field.CanInterface() {
				// Ignore private fields
				continue
			}
			switch field.Interface().(type) {
			// Do not enter inside the time.Time of correct models
			case DateTime, Date:
				return

			case time.Time:
				panic("do not use time.Time in models. Use rdb.Date or rdb.DateTime instead")
			}
			checkFields(field)
		}
	}
}

func (db *Database) Collection(golden Model) *Collection {
	checkFields(reflect.ValueOf(golden))

	return &Collection{
		Query: &Query{
			db:     db,
			conn:   db.conn,
			golden: golden,
			root:   new(andFilter),
		},
		db:     db,
		conn:   db.conn,
		golden: golden,
	}
}

func (db *Database) QueryIndex(index string, golden Model) *Query {
	return &Query{
		db:     db,
		conn:   db.conn,
		golden: golden,
		index:  index,
		root:   new(andFilter),
	}
}

type Index struct {
	Maps              []string
	Reduce            string
	Indexing          map[string]FieldIndexing
	Store             []string
	AdditionalSources []string
}

type FieldIndexing string

const (
	FieldIndexingExact  = "Exact"
	FieldIndexingSearch = "Search"
)

func (db *Database) CreateIndex(ctx context.Context, name string, index Index) error {
	if name == "" {
		return errors.Errorf("index name is required to create it")
	}
	if len(index.Maps) == 0 {
		return errors.Errorf("at least one map is required for a custom index")
	}

	input := &api.Index{
		Type:              "Map",
		Name:              name,
		Maps:              index.Maps,
		AdditionalSources: map[string]string{},
	}
	if index.Reduce != "" {
		input.Type = "MapReduce"
		input.Reduce = &index.Reduce
	}
	input.Fields = make(map[string]*api.IndexFieldOptions)
	for k, v := range index.Indexing {
		input.FieldOrCreate(k).Indexing = string(v)
	}
	for _, field := range index.Store {
		input.FieldOrCreate(field).Storage = "Yes"
	}
	for _, source := range index.AdditionalSources {
		content, err := ioutil.ReadFile(source)
		if err != nil {
			return errors.Trace(err)
		}
		input.AdditionalSources[filepath.Base(source)] = string(content)
	}

	return errors.Trace(db.conn.createIndexes(ctx, []*api.Index{input}))
}

func (db *Database) Patch(ctx context.Context, query *RQLQuery) (*Operation, error) {
	q := &api.Patch{
		Query: &api.Query{
			Query:           query.query,
			QueryParameters: query.queryParams,
		},
	}
	r, err := db.conn.buildPATCH(db.conn.endpoint("queries"), nil, q)
	if err != nil {
		return nil, errors.Trace(err)
	}
	resp, err := db.conn.sendRequest(ctx, r)
	if err != nil {
		return nil, errors.Trace(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, NewUnexpectedStatusError(r, resp)
	}

	op := new(api.Operation)
	if err := json.NewDecoder(resp.Body).Decode(op); err != nil {
		return nil, errors.Trace(err)
	}

	return &Operation{
		conn: db.conn,
		id:   op.ID,
	}, nil
}

func (db *Database) UpsertIdentity(ctx context.Context, name string, value int64) error {
	if name == "" {
		return errors.Errorf("identity name is required to assign it")
	}

	args := map[string]string{
		"name":  name,
		"value": strconv.FormatInt(value, 10),
	}
	r, err := db.conn.buildPOST(db.conn.endpoint("identity/seed"), args, nil)
	if err != nil {
		return errors.Trace(err)
	}

	resp, err := db.conn.sendRequest(ctx, r)
	if err != nil {
		return errors.Trace(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return NewUnexpectedStatusError(r, resp)
	}

	return nil
}

// ConfigureDefaultRevisions for all the collections of this database storing every change.
func (db *Database) ConfigureDefaultRevisions(ctx context.Context, config *api.RevisionConfig) error {
	desc, err := db.conn.descriptor(ctx)
	if err != nil {
		return errors.Trace(err)
	}

	revs := &api.Revisions{
		Default: config,
	}
	if desc.Revisions != nil {
		revs.Collections = desc.Revisions.Collections
	}
	r, err := db.conn.buildPOST(db.conn.endpoint("admin/revisions/config"), nil, revs)
	if err != nil {
		return errors.Trace(err)
	}
	resp, err := db.conn.sendRequest(ctx, r)
	if err != nil {
		return errors.Trace(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return NewUnexpectedStatusError(r, resp)
	}

	return nil
}

func (db *Database) Descriptor(ctx context.Context) (*api.Database, error) {
	return db.conn.descriptor(ctx)
}

func (db *Database) RQLQuery(query *RQLQuery) *DirectQuery {
	return &DirectQuery{
		query: query,
		db:    db,
		conn:  db.conn,
	}
}