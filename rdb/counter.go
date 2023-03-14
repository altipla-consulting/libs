package rdb

import (
	"context"
	"net/http"

	"github.com/altipla-consulting/errors"

	"libs.altipla.consulting/rdb/api"
)

type Counter struct {
	conn        *connection
	docID, name string
	value       int64
}

func (counter *Counter) Value() int64 {
	return counter.value
}

func (counter *Counter) Increment(ctx context.Context, delta int64) error {
	req := &api.CounterOperations{
		Documents: []*api.CounterOperationDocument{
			{
				DocumentID: counter.docID,
				Operations: []*api.CounterOperation{
					{
						CounterName: counter.name,
						Delta:       delta,
						Type:        api.CounterOperationTypeIncrement,
					},
				},
			},
		},
	}

	r, err := counter.conn.buildPOST(counter.conn.endpoint("counters"), nil, req)
	if err != nil {
		return errors.Trace(err)
	}
	resp, err := counter.conn.sendRequest(ctx, r)
	if err != nil {
		var unexpected UnexpectedStatusError
		if errors.As(err, &unexpected) && unexpected.Advanced != nil && unexpected.Advanced.Type == "Raven.Client.Exceptions.Documents.DocumentDoesNotExistException" {
			return newNoSuchEntityError("document %q does not exists when updating counter %q", counter.docID, counter.name)
		}
		return errors.Trace(err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		return nil
	default:
		return NewUnexpectedStatusError(r, resp)
	}
}

func (counter *Counter) Decrement(ctx context.Context, delta int64) error {
	return errors.Trace(counter.Increment(ctx, -1*delta))
}

func (counter *Counter) Delete(ctx context.Context) error {
	req := &api.CounterOperations{
		Documents: []*api.CounterOperationDocument{
			{
				DocumentID: counter.docID,
				Operations: []*api.CounterOperation{
					{
						CounterName: counter.name,
						Type:        api.CounterOperationTypeDelete,
					},
				},
			},
		},
	}

	r, err := counter.conn.buildPOST(counter.conn.endpoint("counters"), nil, req)
	if err != nil {
		return errors.Trace(err)
	}
	resp, err := counter.conn.sendRequest(ctx, r)
	if err != nil {
		return errors.Trace(err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		return nil
	default:
		return NewUnexpectedStatusError(r, resp)
	}
}
