package rdb

import (
	"context"
	"encoding/json"
	"net/http"

	"libs.altipla.consulting/errors"
	"libs.altipla.consulting/naming"
	"libs.altipla.consulting/rdb/api"
)

type Session struct {
	conn *connection

	// Pending actions that will be performed in a single batch
	actions []sessionAction

	includes map[string]api.Result
	counters map[string]int64
}

type sessionAction interface {
	batchCommand() (api.BatchCommand, error)
}

type storeModelAction struct {
	model Model
}

func (action *storeModelAction) batchCommand() (api.BatchCommand, error) {
	id, err := getModelID(action.model)
	if err != nil {
		return nil, errors.Trace(err)
	}
	changeVector := action.model.ChangeVector()
	serialized, err := serializeModel(action.model)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return &api.PutCommand{
		ID:           id,
		ChangeVector: &changeVector,
		Document:     serialized,
		Type:         "PUT",
	}, nil
}

type deleteModelAction struct {
	model Model
}

func (action *deleteModelAction) batchCommand() (api.BatchCommand, error) {
	id, err := getModelID(action.model)
	if err != nil {
		return nil, errors.Trace(err)
	}
	changeVector := action.model.ChangeVector()

	return &api.DeleteCommand{
		ID:           id,
		ChangeVector: &changeVector,
		Type:         "DELETE",
	}, nil
}

type deletePrefixAction struct {
	prefix string
}

func (action *deletePrefixAction) batchCommand() (api.BatchCommand, error) {
	return &api.DeletePrefixCommand{
		ID:         action.prefix,
		IDPrefixed: true,
		Type:       "DELETE",
	}, nil
}

type deleteIDAction struct {
	id string
}

func (action *deleteIDAction) batchCommand() (api.BatchCommand, error) {
	return &api.DeleteCommand{
		ID:   action.id,
		Type: "DELETE",
	}, nil
}

func (sess *Session) SaveChanges(ctx context.Context) error {
	if len(sess.actions) == 0 {
		return nil
	}

	// Optimized single action operations.
	if len(sess.actions) == 1 {
		switch action := sess.actions[0].(type) {
		case *storeModelAction:
			id, err := getModelID(action.model)
			if err != nil {
				return errors.Trace(err)
			}
			params := map[string]string{"id": id}
			serialized, err := serializeModel(action.model)
			if err != nil {
				return errors.Trace(err)
			}
			r, err := sess.conn.buildPUT(sess.conn.endpoint("docs"), params, serialized)
			if err != nil {
				return errors.Trace(err)
			}
			r.Header.Set("If-Match", action.model.ChangeVector())
			resp, err := sess.conn.sendRequest(ctx, r)
			if err != nil {
				return errors.Trace(err)
			}
			defer resp.Body.Close()

			switch resp.StatusCode {
			case http.StatusCreated:
				var metadata api.ModelMetadata
				if err := json.NewDecoder(resp.Body).Decode(&metadata); err != nil {
					return errors.Trace(err)
				}
				action.model.load(metadata.ChangeVector)
				return nil
			case http.StatusConflict:
				return errors.Trace(ErrConcurrentTransaction)
			default:
				return NewUnexpectedStatusError(r, resp)
			}

		case *deleteModelAction:
			id, err := getModelID(action.model)
			if err != nil {
				return errors.Trace(err)
			}
			params := map[string]string{"id": id}
			r, err := sess.conn.buildDELETE(sess.conn.endpoint("docs"), params)
			if err != nil {
				return errors.Trace(err)
			}
			changeVector := action.model.ChangeVector()
			if changeVector == "" {
				return errors.Errorf("retrieve the model before trying to delete it")
			}
			r.Header.Set("If-Match", changeVector)
			resp, err := sess.conn.sendRequest(ctx, r)
			if err != nil {
				return errors.Trace(err)
			}
			defer resp.Body.Close()

			switch resp.StatusCode {
			case http.StatusNoContent:
				action.model.load("")
				return nil
			case http.StatusConflict:
				return errors.Trace(ErrConcurrentTransaction)
			default:
				return NewUnexpectedStatusError(r, resp)
			}

		case *deleteIDAction:
			params := map[string]string{"id": action.id}
			r, err := sess.conn.buildDELETE(sess.conn.endpoint("docs"), params)
			if err != nil {
				return errors.Trace(err)
			}
			resp, err := sess.conn.sendRequest(ctx, r)
			if err != nil {
				return errors.Trace(err)
			}
			defer resp.Body.Close()

			switch resp.StatusCode {
			case http.StatusNoContent:
				return nil
			default:
				return NewUnexpectedStatusError(r, resp)
			}
		}
	}

	// Batch commands sending multiple operations in one go.
	batch := new(api.BulkCommands)
	for _, action := range sess.actions {
		cmd, err := action.batchCommand()
		if err != nil {
			return errors.Trace(err)
		}
		batch.Commands = append(batch.Commands, cmd)
	}

	r, err := sess.conn.buildPOST(sess.conn.endpoint("bulk_docs"), nil, batch)
	if err != nil {
		return errors.Trace(err)
	}
	resp, err := sess.conn.sendRequest(ctx, r)
	if err != nil {
		return errors.Trace(err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusCreated:
		results := new(api.Results)
		if err := json.NewDecoder(resp.Body).Decode(results); err != nil {
			return errors.Trace(err)
		}
		for i, result := range results.Results {
			if store, ok := sess.actions[i].(*storeModelAction); ok {
				store.model.load(result.DirectMetadata("@change-vector"))
			}
		}
	case http.StatusConflict:
		return errors.Trace(ErrConcurrentTransaction)
	default:
		return NewUnexpectedStatusError(r, resp)
	}

	sess.actions = nil
	return nil
}

func (sess *Session) Load(id string, dest interface{}) error {
	result, ok := sess.includes[id]
	if !ok {
		return errors.Wrapf(ErrNoSuchEntity, "included id: %s", id)
	}
	_, err := createModel(dest, result)
	return errors.Trace(err)
}

func (sess *Session) Counter(docID string, name string) *Counter {
	return &Counter{
		conn:  sess.conn,
		docID: docID,
		name:  name,
		value: sess.counters[naming.Generate(docID, name)],
	}
}

func (sess *Session) mergeIncludes(includes map[string]api.Result) {
	if sess == nil {
		return
	}
	if sess.includes == nil {
		sess.includes = make(map[string]api.Result)
	}
	for k, v := range includes {
		sess.includes[k] = v
	}
}

func (sess *Session) mergeCounters(counters *api.Counters) {
	if sess.counters == nil {
		sess.counters = make(map[string]int64)
	}
	for _, counter := range counters.Counters {
		sess.counters[naming.Generate(counter.DocumentID, counter.CounterName)] = counter.TotalValue
	}
}
