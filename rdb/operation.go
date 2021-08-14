package rdb

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"libs.altipla.consulting/errors"
	"libs.altipla.consulting/rdb/api"
)

type Operation struct {
	conn *connection
	id   int64
}

type OperationProgress struct {
	Processed int64
	Total     int64
}

func (op *Operation) Progress(ctx context.Context) (OperationProgress, error) {
	st, err := op.fetchStatus(ctx)
	if err != nil {
		return OperationProgress{}, errors.Trace(err)
	}
	if st.Progress != nil {
		return OperationProgress{
			Processed: st.Progress.Processed,
			Total:     st.Progress.Total,
		}, nil
	}
	if st.Result != nil {
		return OperationProgress{
			Processed: st.Result.Total,
			Total:     st.Result.Total,
		}, nil
	}
	return OperationProgress{}, nil
}

func (op *Operation) WaitFor(ctx context.Context, backoff time.Duration) error {
	for {
		if ctx.Err() != nil {
			return errors.Trace(ctx.Err())
		}

		status, err := op.fetchStatus(ctx)
		if err != nil {
			return errors.Trace(err)
		} else if status.Status == api.OperationStatusCompleted {
			break
		}

		time.Sleep(backoff)
	}

	return nil
}

func (op *Operation) fetchStatus(ctx context.Context) (*api.OperationStatus, error) {
	qs := map[string]interface{}{"id": strconv.FormatInt(op.id, 10)}
	r, err := op.conn.buildGET(op.conn.endpoint("operations/state"), qs)
	if err != nil {
		return nil, errors.Trace(err)
	}
	resp, err := op.conn.sendRequest(ctx, r)
	if err != nil {
		return nil, errors.Trace(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, NewUnexpectedStatusError(r, resp)
	}

	status := new(api.OperationStatus)
	if err := json.NewDecoder(resp.Body).Decode(status); err != nil {
		return nil, errors.Trace(err)
	}

	if status.Status == api.OperationStatusFaulted && status.Result != nil {
		return nil, errors.Wrapf(errors.Errorf("operation failed: %s", status.Result.Message), "database side stack:\n%s", status.Result.Error)
	}

	return status, nil
}
