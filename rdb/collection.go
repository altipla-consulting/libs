package rdb

import (
	"context"
	"encoding/json"
	"net/http"
	"reflect"

	"libs.altipla.consulting/errors"
	"libs.altipla.consulting/rdb/api"
)

type Collection struct {
	*Query
	db        *Database
	conn      *connection
	golden    Model
	enforcers []ModelEnforcer
}

type ModelEnforcer func(model Model) bool
type QueryEnforcer func(q *Query) *Query

type Enforcer struct {
	Model ModelEnforcer
	Query QueryEnforcer
}

func (collection *Collection) Enforce(enforcer Enforcer) *Collection {
	if enforcer.Model == nil {
		panic("enforcer.Model cannot be nil")
	}
	if enforcer.Query == nil {
		panic("enforcer.Query cannot be nil")
	}

	collection.enforcers = append(collection.enforcers, enforcer.Model)
	collection.Query.enforcers = append(collection.Query.enforcers, enforcer.Query)
	return collection
}

func (collection *Collection) checkEnforcers(model Model) bool {
	for _, fn := range collection.enforcers {
		if !fn(model) {
			return false
		}
	}
	return true
}

func (collection *Collection) Put(ctx context.Context, model Model) error {
	if !collection.checkEnforcers(model) {
		return errors.Wrapf(ErrNoSuchEntity, "enforced model")
	}

	sess := SessionFromContext(ctx)
	if sess == nil {
		ctx, sess := collection.db.NewSession(ctx)
		sess.actions = append(sess.actions, &storeModelAction{model})
		return errors.Trace(sess.SaveChanges(ctx))
	}

	sess.actions = append(sess.actions, &storeModelAction{model})
	return nil
}

// Get fetchs an entity by its ID. If it doesn't exists or it's enforced
// it returns ErrNoSuchEntity.
func (collection *Collection) Get(ctx context.Context, id string, dest interface{}, opts ...IncludeOption) error {
	sess := SessionFromContext(ctx)
	if sess == nil && len(opts) > 0 {
		return errors.Errorf("cannot include additional entities without a session")
	}

	if id == "" {
		return errors.Wrapf(ErrNoSuchEntity, "empty id")
	}
	if l := len([]byte(id)); l >= 510 {
		return errors.Wrapf(ErrNoSuchEntity, "id is too long to be valid: %v", l)
	}

	params := map[string]interface{}{
		"id":      id,
		"include": applyModelIncludes(opts...),
	}
	r, err := collection.conn.buildGET(collection.conn.endpoint("docs"), params)
	if err != nil {
		return errors.Trace(err)
	}
	resp, err := collection.conn.sendRequest(ctx, r)
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
		sess.mergeIncludes(results.Includes)
		model, err := createModel(dest, results.Results[0])
		if err != nil {
			return errors.Trace(err)
		}
		if model != nil {
			if !collection.checkEnforcers(model) {
				// Try to avoid silly errors cleaning up the model to be nil even if
				// we are returning an error.
				if err := json.Unmarshal([]byte("null"), dest); err != nil {
					return errors.Trace(err)
				}
				return errors.Wrapf(ErrNoSuchEntity, "enforced id: %s", id)
			}
		} else if len(collection.enforcers) > 0 {
			return errors.Errorf("cannot enforce non-model entities")
		}
	case http.StatusNotFound:
		return errors.Wrapf(ErrNoSuchEntity, "id: %s", id)
	default:
		return NewUnexpectedStatusError(r, resp)
	}

	if resolveIncludes(opts...).allCounters {
		params = map[string]interface{}{"docId": id}
		r, err = collection.conn.buildGET(collection.conn.endpoint("counters"), params)
		if err != nil {
			return errors.Trace(err)
		}
		resp, err := collection.conn.sendRequest(ctx, r)
		if err != nil {
			return errors.Trace(err)
		}
		defer resp.Body.Close()

		switch resp.StatusCode {
		case http.StatusOK:
			counters := new(api.Counters)
			if err := json.NewDecoder(resp.Body).Decode(counters); err != nil {
				return errors.Trace(err)
			}
			sess.mergeCounters(counters)
		default:
			return NewUnexpectedStatusError(r, resp)
		}
	}

	return nil
}

// TryGet tries to fetch the entity, but returns successfully even if it doesn't exist.
// The model will be untouched in case it doesn't exists.
func (collection *Collection) TryGet(ctx context.Context, id string, model interface{}, opts ...IncludeOption) error {
	if err := collection.Get(ctx, id, model, opts...); err != nil {
		if errors.Is(err, ErrNoSuchEntity) {
			return nil
		}

		return errors.Trace(err)
	}
	return nil
}

// GetMulti fetches multiple models. If one of them doesn't exists it will return ErrNoSuchEntity.
func (collection *Collection) GetMulti(ctx context.Context, ids []string, dest interface{}, opts ...IncludeOption) error {
	rt := reflect.TypeOf(dest)
	if rt.Kind() != reflect.Ptr || rt.Elem().Kind() != reflect.Slice || rt.Elem().Elem().Kind() != reflect.Ptr || rt.Elem().Elem().Elem().Kind() != reflect.Struct {
		return errors.Errorf("dest should be a pointer to a slice of models: %T", dest)
	}

	sess := SessionFromContext(ctx)
	if sess == nil && len(opts) > 0 {
		return errors.Errorf("cannot include additional entities without a session")
	}

	if len(ids) == 0 {
		return nil
	}

	params := map[string]interface{}{
		"include": applyModelIncludes(opts...),
	}
	body := &api.DocsRequest{IDs: ids}
	r, err := collection.conn.buildPOST(collection.conn.endpoint("docs"), params, body)
	if err != nil {
		return errors.Trace(err)
	}
	resp, err := collection.conn.sendRequest(ctx, r)
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
		sess.mergeIncludes(results.Includes)

		merr := make(MultiError, len(ids))
		slice := reflect.MakeSlice(rt.Elem(), 0, len(ids))
		for i, result := range results.Results {
			if result == nil {
				merr[i] = errors.Wrapf(ErrNoSuchEntity, "id: %s", ids[i])
				slice = reflect.Append(slice, reflect.Zero(rt.Elem().Elem()))
				continue
			}
			item := reflect.New(reflect.PtrTo(rt.Elem().Elem().Elem()))
			model, err := createModel(item.Interface(), result)
			if err != nil {
				merr[i] = errors.Trace(err)
			}
			if model != nil {
				if !collection.checkEnforcers(model) {
					merr[i] = errors.Wrapf(ErrNoSuchEntity, "enforced id: %s", ids[i])
					slice = reflect.Append(slice, reflect.Zero(rt.Elem().Elem()))
					continue
				}
			} else if len(collection.enforcers) > 0 {
				return errors.Errorf("cannot enforce non-model entities")
			}
			slice = reflect.Append(slice, item.Elem())
		}
		reflect.ValueOf(dest).Elem().Set(slice)
		if merr.HasError() {
			return errors.Trace(merr)
		}

		return nil
	case http.StatusNotFound:
		return errors.Wrapf(ErrNoSuchEntity, "ids: %s", ids)
	default:
		return NewUnexpectedStatusError(r, resp)
	}
}

func (collection *Collection) Delete(ctx context.Context, model Model) error {
	if !collection.checkEnforcers(model) {
		return errors.Wrapf(ErrNoSuchEntity, "enforced model")
	}

	sess := SessionFromContext(ctx)
	if sess == nil {
		ctx, sess := collection.db.NewSession(ctx)
		sess.actions = append(sess.actions, &deleteModelAction{model})
		return errors.Trace(sess.SaveChanges(ctx))
	}

	sess.actions = append(sess.actions, &deleteModelAction{model})
	return nil
}

func (collection *Collection) Exists(ctx context.Context, id string) (bool, error) {
	if len(collection.enforcers) > 0 {
		return false, errors.Errorf("cannot enforce entities while calling Exists")
	}
	if id == "" {
		return false, errors.Wrapf(ErrNoSuchEntity, "empty id")
	}

	params := map[string]interface{}{
		"id":           id,
		"metadataOnly": "true",
	}
	r, err := collection.conn.buildGET(collection.conn.endpoint("docs"), params)
	if err != nil {
		return false, errors.Trace(err)
	}
	resp, err := collection.conn.sendRequest(ctx, r)
	if err != nil {
		return false, errors.Trace(err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		return true, nil
	case http.StatusNotFound:
		return false, nil
	default:
		return false, NewUnexpectedStatusError(r, resp)
	}
}

// ConfigureRevisions for this collection to store every change.
func (collection *Collection) ConfigureRevisions(ctx context.Context, config *api.RevisionConfig) error {
	desc, err := collection.conn.descriptor(ctx)
	if err != nil {
		return errors.Trace(err)
	}

	revs := &api.Revisions{
		Collections: map[string]*api.RevisionConfig{},
	}
	if config != nil {
		revs.Collections[collection.golden.Collection()] = config
	}
	if desc.Revisions != nil {
		revs.Default = desc.Revisions.Default
		for k, v := range desc.Revisions.Collections {
			if k != collection.golden.Collection() {
				revs.Collections[k] = v
			}
		}
	}
	r, err := collection.conn.buildPOST(collection.conn.endpoint("admin/revisions/config"), nil, revs)
	if err != nil {
		return errors.Trace(err)
	}
	resp, err := collection.conn.sendRequest(ctx, r)
	if err != nil {
		return errors.Trace(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return NewUnexpectedStatusError(r, resp)
	}

	return nil
}
