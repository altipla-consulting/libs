package bigquery

import (
	"context"
	"encoding/base64"
	"reflect"

	"cloud.google.com/go/bigquery"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	pb "libs.altipla.consulting/bigquery/proto"
	"libs.altipla.consulting/errors"
)

const (
	// DefaultPageSize if not specified.
	DefaultPageSize = 50

	// MaxPageSize we can ask for in the pagination.
	MaxPageSize = 500
)

// Pager helps when retrieving paginated results.
type Pager struct {
	// NextPageToken is the token of the next page, or empty if there is no more
	// results. It is filled after the call to Fetch.
	NextPageToken string

	// TotalSize is the total number of results the job has returned. It is filled
	// after the call to Fetch.
	TotalSize int32

	bq        *bigquery.Client
	query     *Query
	pageSize  int32
	pageToken string
	dataset   Dataset
}

// SetInputs fills the pagination inputs we receive from the client.
func (pager *Pager) SetInputs(pageToken string, pageSize int32) {
	pager.pageToken = pageToken
	pager.pageSize = pageSize
	if pager.pageSize <= 0 {
		pager.pageSize = DefaultPageSize
	}
	if pager.pageSize > MaxPageSize {
		pager.pageSize = MaxPageSize
	}
}

// Fetch retrieves the next page of results. Pass a pointer to an empty slice
// of the type of model you need to read.
func (pager *Pager) Fetch(ctx context.Context, models interface{}) error {
	// It is safe to convert the uint32 to a int64 and we will
	// never get a negative number doing so.
	checksum := int64(pager.query.Checksum(pager.dataset, pager.pageSize))

	var it *bigquery.RowIterator
	var jobID string
	if pager.pageToken == "" {
		b := new(sqlBuilder)
		q := pager.bq.Query(pager.query.buildSQL(pager.dataset, b))
		q.Parameters = b.params

		job, err := q.Run(ctx)
		if err != nil {
			return errors.Trace(err)
		}
		jobID = job.ID()

		it, err = job.Read(ctx)
		if err != nil {
			return errors.Trace(err)
		}

		// Cuando no hay resultados en la query it.PageInfo() es nulo y provocaría un pánico al modificar el token.
		if it.TotalRows == 0 {
			return nil
		}
	} else {
		raw, err := base64.StdEncoding.DecodeString(pager.pageToken)
		if err != nil {
			return errors.Trace(err)
		}
		token := new(pb.Token)
		if err := proto.Unmarshal(raw, token); err != nil {
			return errors.Trace(err)
		}

		if checksum != token.Checksum {
			return status.Errorf(codes.InvalidArgument, "invalid pagination token: wrong checksum: %v", pager.pageToken)
		}

		job, err := pager.bq.JobFromID(ctx, token.JobId)
		if err != nil {
			return errors.Trace(err)
		}
		jobID = job.ID()

		it, err = job.Read(ctx)
		if err != nil {
			return errors.Trace(err)
		}

		it.PageInfo().Token = token.PageToken
	}

	vt := reflect.TypeOf(models)

	if vt.Kind() != reflect.Ptr && vt.Elem().Kind() != reflect.Slice {
		return errors.Errorf("pass a pointer to a slice to Fetch")
	}
	if vt.Elem().Elem().Kind() != reflect.Ptr || vt.Elem().Elem().Elem().Kind() != reflect.Struct {
		return errors.Errorf("pass a pointer to a slice of struct pointers to Fetch")
	}

	it.PageInfo().MaxSize = int(pager.pageSize)
	dest := reflect.MakeSlice(vt.Elem(), 0, 0)
	for {
		model := reflect.New(vt.Elem().Elem().Elem())
		if err := it.Next(model.Interface()); err != nil {
			if errors.Is(err, iterator.Done) {
				break
			}
			return errors.Trace(err)
		}

		dest = reflect.Append(dest, model)
		if dest.Len() == int(pager.pageSize) {
			break
		}
	}
	reflect.ValueOf(models).Elem().Set(dest)

	pager.TotalSize = int32(it.TotalRows)
	pager.NextPageToken = ""

	if it.PageInfo().Token != "" {
		token := &pb.Token{
			JobId:     jobID,
			PageToken: it.PageInfo().Token,
			Checksum:  checksum,
		}
		raw, err := proto.Marshal(token)
		if err != nil {
			return errors.Trace(err)
		}
		pager.NextPageToken = base64.StdEncoding.EncodeToString(raw)
	}

	return nil
}
