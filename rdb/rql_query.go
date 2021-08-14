package rdb

import (
	"context"
	"encoding/json"
	"hash/crc32"
	"net/http"
	"reflect"

	"libs.altipla.consulting/errors"
	"libs.altipla.consulting/rdb/api"
)

type RQLQuery struct {
	query       string
	queryParams map[string]interface{}
}

func NewRQLQuery(query string) *RQLQuery {
	return &RQLQuery{
		query:       query,
		queryParams: make(map[string]interface{}),
	}
}

func (q *RQLQuery) Clone() *RQLQuery {
	params := make(map[string]interface{})
	for k, v := range q.queryParams {
		params[k] = v
	}

	return &RQLQuery{
		query:       q.query,
		queryParams: q.queryParams,
	}
}

func (q *RQLQuery) Set(param string, value interface{}) {
	q.queryParams[param] = value
}

type DirectQuery struct {
	query *RQLQuery
	db    *Database
	conn  *connection

	strongConsistency bool

	// Stats of the last server operation
	stats QueryStats
}

func (q *DirectQuery) Clone() *DirectQuery {
	return &DirectQuery{
		db:                q.db,
		conn:              q.conn,
		query:             q.query.Clone(),
		strongConsistency: q.strongConsistency,
	}
}

func (q *DirectQuery) String() string {
	rql, _ := q.RQL()
	return rql
}

func (q *DirectQuery) RQL() (string, map[string]interface{}) {
	return q.query.query, q.query.queryParams
}

func (q *DirectQuery) ForceStrongConsistency() *DirectQuery {
	q.strongConsistency = true
	return q
}

// Stats return general statistics about the query AFTER it is performed. Calling
// it before retrieving results would return a nil value.
func (q *DirectQuery) Stats() QueryStats {
	return q.stats
}

func (q *DirectQuery) setStats(results *api.Results) {
	q.stats = QueryStats{
		TotalResults:   results.TotalResults,
		SkippedResults: results.SkippedResults,
		DurationInMs:   results.DurationInMs,
		IsStale:        results.IsStale,
		IndexName:      results.IndexName,
		ResultSize:     results.TotalResults,
	}
	if results.CappedMaxResults != nil {
		q.stats.ResultSize = *results.CappedMaxResults
	}
}

func (q *DirectQuery) GetAll(ctx context.Context, dest interface{}) error {
	rt := reflect.TypeOf(dest)
	if rt.Kind() != reflect.Ptr || rt.Elem().Kind() != reflect.Slice || rt.Elem().Elem().Kind() != reflect.Ptr || rt.Elem().Elem().Elem().Kind() != reflect.Struct {
		return errors.Errorf("dest should be a pointer to a slice of models: %T", dest)
	}

	sess := SessionFromContext(ctx)
	if sess == nil {
		return errors.Errorf("direct queries always need a session to fetch includes")
	}

	r, err := q.conn.buildPOST(q.conn.endpoint("queries"), nil, q.buildQuery())
	if err != nil {
		return errors.Trace(err)
	}
	resp, err := q.conn.sendRequest(ctx, r)
	if err != nil {
		return errors.Trace(err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		results := new(api.Results)
		if err := json.NewDecoder(resp.Body).Decode(results); err != nil {
			return errors.Trace(err)
		}
		q.setStats(results)
		sess.mergeIncludes(results.Includes)
		metadata := make([]ModelMetadata, len(results.Results))
		for i, result := range results.Results {
			metadata[i] = ModelMetadata{
				ID:           result.Metadata("@id"),
				ChangeVector: result.Metadata("@change-vector"),
			}
		}
		slice := reflect.MakeSlice(rt.Elem(), 0, len(results.Results))
		for _, result := range results.Results {
			item := reflect.New(reflect.PtrTo(rt.Elem().Elem().Elem()))
			if _, err := createModel(item.Interface(), result); err != nil {
				return errors.Trace(err)
			}
			slice = reflect.Append(slice, item.Elem())
		}
		reflect.ValueOf(dest).Elem().Set(slice)
		return nil
	default:
		return NewUnexpectedStatusError(r, resp)
	}
}

func (q *DirectQuery) buildQuery() *api.Query {
	rql, params := q.RQL()
	query := &api.Query{
		Query:                  rql,
		QueryParameters:        params,
		WaitForNonStaleResults: q.strongConsistency || q.db.strongConsistency,
	}
	if query.WaitForNonStaleResults {
		query.WaitForNonStaleResultsTimeout = "00:00:15"
	}
	return query
}

// Checksum returns a checksum of the filters, conditions, table name, ... and other
// internal data of the collection that identifies it.
func (q *DirectQuery) Checksum() uint32 {
	rql, params := q.RQL()
	encoded, err := json.Marshal(params)
	if err != nil {
		panic(err)
	}
	return crc32.ChecksumIEEE([]byte(rql + string(encoded)))
}
