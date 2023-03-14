package rdb

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"net/http"
	"reflect"
	"strings"

	"github.com/altipla-consulting/errors"

	"libs.altipla.consulting/rdb/api"
)

type Query struct {
	db        *Database
	conn      *connection
	index     string
	golden    Model
	enforcers []QueryEnforcer

	strongConsistency bool
	statsOnly         bool
	offset, limit     int64
	orders            []string
	randomOrder       bool
	root              *andFilter
	selectGolden      interface{}
	selectFields      []string
	includes          []string

	// Stats of the last server operation
	stats QueryStats
}

func (q *Query) Clone() *Query {
	return &Query{
		db:                q.db,
		conn:              q.conn,
		index:             q.index,
		golden:            q.golden,
		enforcers:         q.enforcers,
		strongConsistency: q.strongConsistency,
		statsOnly:         q.statsOnly,
		offset:            q.offset,
		limit:             q.limit,
		orders:            q.orders,
		randomOrder:       q.randomOrder,
		root:              q.root.clone(),
		selectGolden:      q.selectGolden,
		selectFields:      q.selectFields,
		includes:          q.includes,
	}
}

func (q *Query) String() string {
	rql, _ := q.RQL()
	return rql
}

func (q *Query) RQL() (string, map[string]interface{}) {
	q = q.Clone()
	for _, fn := range q.enforcers {
		q = fn(q)
	}

	parts := []string{}
	params := NewParams()

	if q.index != "" {
		parts = append(parts, "from index '"+q.index+"'")
	} else {
		parts = append(parts, "from "+q.golden.Collection())
	}
	if !q.root.isEmpty() {
		var rql string
		rql = q.root.RQL(params)
		parts = append(parts, "where "+rql)
	}
	if len(q.orders) > 0 {
		for i, order := range q.orders {
			if strings.HasPrefix(order, "-") {
				q.orders[i] = order[1:] + " desc"
			}
		}
		parts = append(parts, "order by "+strings.Join(q.orders, ", "))
	}
	if q.randomOrder {
		parts = append(parts, "order by random()")
	}
	if len(q.selectFields) > 0 {
		parts = append(parts, "select "+strings.Join(q.selectFields, ", "))
	}
	if len(q.includes) > 0 {
		parts = append(parts, "include "+strings.Join(q.includes, ", "))
	}
	if q.limit > 0 {
		parts = append(parts, fmt.Sprintf("limit %d", q.limit))
	}
	if q.statsOnly {
		parts = append(parts, "limit 0")
	}
	if q.offset > 0 {
		parts = append(parts, fmt.Sprintf("offset %d", q.offset))
	}

	return strings.Join(parts, " "), params.values
}

func (q *Query) Offset(offset int64) *Query {
	q.offset = offset
	return q
}

func (q *Query) Limit(limit int64) *Query {
	q.limit = limit
	return q
}

func (q *Query) checkOrderBy(field string) {
	if q.randomOrder {
		panic("cannot use OrderBy after RandomOrder")
	}
	if strings.Contains(field, " ") {
		panic("do not use OrderBy with different columns, use multiple calls")
	}
	if strings.Contains(field, ",") {
		panic("do not use OrderBy with different columns, use multiple calls")
	}
}

func (q *Query) OrderBy(field string) *Query {
	rt := reflect.TypeOf(q.golden)

	fieldName := field
	if strings.HasPrefix(fieldName, "-") {
		fieldName = fieldName[1:]
	}
	if _, ok := rt.Elem().FieldByName(fieldName); !ok {
		panic("OrderBy cannot find the field specified: " + fieldName)
	}
	example := reflect.ValueOf(q.golden).Elem().FieldByName(fieldName).Interface()
	switch example.(type) {
	case string, Date, DateTime, bool:
		return q.OrderByAlpha(field)
	case int64, int32, int:
		return q.OrderByNumeric(field)
	}
	panic(fmt.Sprintf("cannot detect field type in OrderBy: %s: %T", fieldName, example))
}

func (q *Query) OrderByAlpha(field string) *Query {
	q.checkOrderBy(field)
	q.orders = append(q.orders, field+" as string")
	return q
}

func (q *Query) OrderByAlphaNumeric(field string) *Query {
	q.checkOrderBy(field)
	q.orders = append(q.orders, field+" as alphanumeric")
	return q
}

func (q *Query) OrderByNumeric(field string) *Query {
	q.checkOrderBy(field)
	q.orders = append(q.orders, field+" as long")
	return q
}

func (q *Query) RandomOrder() *Query {
	if len(q.orders) > 0 {
		panic("cannot use RandomOrder after OrderBy")
	}
	q.randomOrder = true
	return q
}

func (q *Query) Filter(field string, value interface{}) *Query {
	q.root.children = append(q.root.children, Filter(field, value))
	return q
}

func (q *Query) FilterNotExact(field string, value interface{}) *Query {
	q.root.children = append(q.root.children, FilterNotExact(field, value))
	return q
}

func (q *Query) FilterSub(filters ...QueryFilter) *Query {
	q.root.children = append(q.root.children, filters...)
	return q
}

func (q *Query) FilterIn(field string, values ...interface{}) *Query {
	q.root.children = append(q.root.children, FilterIn(field, values...))
	return q
}

func (q *Query) FilterContainsAll(field string, values ...interface{}) *Query {
	q.root.children = append(q.root.children, FilterContainsAll(field, values...))
	return q
}

func (q *Query) FilterStartsWith(field, prefix string) *Query {
	q.root.children = append(q.root.children, FilterStartsWith(field, prefix))
	return q
}

func (q *Query) FilterEndsWith(field, suffix string) *Query {
	q.root.children = append(q.root.children, FilterEndsWith(field, suffix))
	return q
}

func EscapeSearch(str string) string {
	return strings.Replace(str, "*", `\*`, -1)
}

type SearchOption string

const SearchOptionAnd = SearchOption("and")
const SearchOptionOr = SearchOption("or")

func (q *Query) FilterSearch(field, search string, opts ...SearchOption) *Query {
	q.root.children = append(q.root.children, FilterSearch(field, search, opts...))
	return q
}

func (q *Query) FilterHasField(field string) *Query {
	q.root.children = append(q.root.children, FilterHasField(field))
	return q
}

func (q *Query) FilterBetween(field string, start, end interface{}) *Query {
	q.root.children = append(q.root.children, FilterBetween(field, start, end))
	return q
}

func (q *Query) ForceStrongConsistency() *Query {
	q.strongConsistency = true
	return q
}

type QueryStats struct {
	// Total number of results of the query or collection ignoring pagination.
	TotalResults int64

	// Number of fake results skipped server side. This occurs during Distinct
	// queries and fan out indexes only.
	SkippedResults int64

	// Duration server side in milliseconds.
	DurationInMs int64

	// If the results are from an index, this flag tells us if the results
	// might be stale because the index was being rebuilt.
	IsStale bool

	// Name of the index or collection used to extract the results.
	IndexName string

	// Number of results of this page, or TotalResults if pagination
	// with Limit and Offset was not applied.
	ResultSize int64
}

// Stats return general statistics about the query AFTER it is performed. Calling
// it before retrieving results would return a nil value.
func (q *Query) Stats() QueryStats {
	return q.stats
}

func (q *Query) setStats(results *api.Results) {
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

func (q *Query) GetAll(ctx context.Context, dest interface{}, opts ...IncludeOption) error {
	rt := reflect.TypeOf(dest)
	if rt.Kind() != reflect.Ptr || rt.Elem().Kind() != reflect.Slice || rt.Elem().Elem().Kind() != reflect.Ptr || rt.Elem().Elem().Elem().Kind() != reflect.Struct {
		return errors.Errorf("dest should be a pointer to a slice of models: %T", dest)
	}

	sess := SessionFromContext(ctx)
	if sess == nil && len(opts) > 0 {
		return errors.Errorf("cannot include additional entities without a session")
	}

	r, err := q.conn.buildPOST(q.conn.endpoint("queries"), nil, q.buildQuery(opts...))
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
	case http.StatusNotFound:
		return errors.Errorf("index not found: %s", q.index)
	default:
		return NewUnexpectedStatusError(r, resp)
	}
}

func (q *Query) Count(ctx context.Context) (int64, error) {
	params := map[string]interface{}{"metadataOnly": "true"}
	q.statsOnly = true
	r, err := q.conn.buildPOST(q.conn.endpoint("queries"), params, q.buildQuery())
	if err != nil {
		return 0, errors.Trace(err)
	}
	resp, err := q.conn.sendRequest(ctx, r)
	if err != nil {
		return 0, errors.Trace(err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		results := new(api.Results)
		if err := json.NewDecoder(resp.Body).Decode(results); err != nil {
			return 0, errors.Trace(err)
		}
		q.setStats(results)
		return results.TotalResults, nil
	default:
		return 0, NewUnexpectedStatusError(r, resp)
	}
}

func (q *Query) HasResults(ctx context.Context) (bool, error) {
	count, err := q.Count(ctx)
	if err != nil {
		return false, errors.Trace(err)
	}
	return count > 0, nil
}

func (q *Query) First(ctx context.Context, dest interface{}, opts ...IncludeOption) error {
	if err := checkSingleModel(dest); err != nil {
		return errors.Trace(err)
	}

	sess := SessionFromContext(ctx)
	if sess == nil && len(opts) > 0 {
		return errors.Errorf("cannot include additional entities without a session")
	}

	r, err := q.conn.buildPOST(q.conn.endpoint("queries"), nil, q.Limit(1).buildQuery(opts...))
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
		if len(results.Results) > 0 {
			if _, err := createModel(dest, results.Results[0]); err != nil {
				return errors.Trace(err)
			}
		}
		return nil
	case http.StatusNotFound:
		return errors.Errorf("index not found: %s", q.index)
	default:
		return NewUnexpectedStatusError(r, resp)
	}
}

func (q *Query) buildQuery(opts ...IncludeOption) *api.Query {
	q.includes = applyModelIncludes(opts...)
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

type ModelMetadata = api.ModelMetadata

func (q *Query) GetAllMetadata(ctx context.Context) ([]ModelMetadata, error) {
	params := map[string]interface{}{"metadataOnly": "true"}
	r, err := q.conn.buildPOST(q.conn.endpoint("queries"), params, q.buildQuery())
	if err != nil {
		return nil, errors.Trace(err)
	}
	resp, err := q.conn.sendRequest(ctx, r)
	if err != nil {
		return nil, errors.Trace(err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		results := new(api.Results)
		if err := json.NewDecoder(resp.Body).Decode(results); err != nil {
			return nil, errors.Trace(err)
		}
		q.setStats(results)
		metadata := make([]ModelMetadata, len(results.Results))
		for i, result := range results.Results {
			metadata[i] = ModelMetadata{
				ID:           result.Metadata("@id"),
				ChangeVector: result.Metadata("@change-vector"),
			}
		}
		return metadata, nil
	case http.StatusNotFound:
		return nil, errors.Errorf("index not found: %s", q.index)
	default:
		return nil, NewUnexpectedStatusError(r, resp)
	}
}

func (q *Query) GetAllIDs(ctx context.Context) ([]string, error) {
	metadata, err := q.GetAllMetadata(ctx)
	if err != nil {
		return nil, errors.Trace(err)
	}
	ids := make([]string, len(metadata))
	for i, md := range metadata {
		ids[i] = md.ID
	}
	return ids, nil
}

func (q *Query) DeleteEverything(ctx context.Context) error {
	// Is much quicker, though a dirty trick, to remove everything by its ID
	// manually instead of using the Delete global operation. We use this method
	// in every test so it's an important optimization.

	metadata, err := q.GetAllMetadata(ctx)
	if err != nil {
		return errors.Trace(err)
	}
	ids := make([]string, len(metadata))
	for i, md := range metadata {
		ids[i] = md.ID
	}

	sess := SessionFromContext(ctx)
	if sess == nil {
		ctx, sess := q.db.NewSession(ctx)
		for _, md := range metadata {
			sess.actions = append(sess.actions, &deleteIDAction{md.ID})
		}
		return errors.Trace(sess.SaveChanges(ctx))
	}

	for _, md := range metadata {
		sess.actions = append(sess.actions, &deleteIDAction{md.ID})
	}
	return nil
}

func (q *Query) Select(fields ...string) *Query {
	if len(fields) == 0 {
		panic("cannot select zero fields")
	}
	if len(q.selectFields) > 0 {
		panic("cannot select fields twice")
	}
	rt := reflect.TypeOf(q.golden).Elem()
	for _, field := range fields {
		if strings.ToLower(field) == "id" {
			panic("ID cannot be projected, use a metadata only query with GetAllMetadata or GetAllIDs")
		}
		if _, ok := rt.FieldByName(field); !ok {
			panic("field " + field + " was not found in the collection model")
		}
	}

	q.selectGolden = q.golden
	q.selectFields = fields
	return q
}

func (q *Query) Project(golden interface{}, fields ...string) *Query {
	if len(fields) == 0 {
		panic("cannot project zero fields")
	}
	if len(q.selectFields) > 0 {
		panic("cannot project fields twice")
	}
	rt := reflect.TypeOf(golden).Elem()
	for _, field := range fields {
		if field == "ID" {
			panic("ID cannot be projected, use a metadata only query with GetAllMetadata or GetAllIDs")
		}
		if _, ok := rt.FieldByName(field); !ok {
			panic("field " + field + " was not found in the projected model")
		}
	}

	q.selectGolden = golden
	q.selectFields = fields
	return q
}

// Checksum returns a checksum of the filters, conditions, table name, ... and other
// internal data of the collection that identifies it.
func (q *Query) Checksum() uint32 {
	rql, params := q.RQL()
	encoded, err := json.Marshal(params)
	if err != nil {
		panic(err)
	}
	return crc32.ChecksumIEEE([]byte(rql + string(encoded)))
}
