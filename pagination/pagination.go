package pagination

import (
	"context"
	"reflect"

	"github.com/speps/go-hashids"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"libs.altipla.consulting/database"
	"libs.altipla.consulting/errors"
)

const (
	DefaultPageSize = 100
	MaxPageSize = 1000

	Alphabet = "abcdefghijklmnopqrstuvwxyz1234567890"
)

type Pager struct {
	NextPageToken string
	TotalSize     int32
	PrevPageToken string

	c         *database.Collection
	pageSize  int32
	pageToken string
}

func NewPager(c *database.Collection) *Pager {
	return &Pager{
		c: c,
	}
}

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

func (pager *Pager) Fetch(ctx context.Context, models interface{}) error {
	// Count the page size between the params we checksum.
	c := pager.c.Clone().Limit(int64(pager.pageSize))

	// It is safe to convert the uint32 to a int64 and we will
	// never get a negative number doing so.
	checksum := int64(c.Checksum())

	hd := hashids.NewData()
	hd.Alphabet = Alphabet
	hd.Salt = "libs.altipla.consulting/pagination"
	h, err := hashids.NewWithData(hd)
	if err != nil {
		return errors.Trace(err)
	}

	var start int64
	if pager.pageToken != "" {
		decoded, err := h.DecodeInt64WithError(pager.pageToken)
		if err != nil {
			return status.Errorf(codes.InvalidArgument, "invalid pagination token: %v: %v", pager.pageToken, err)
		}

		start = decoded[1]
		if start < 0 {
			start = 0
		}

		if checksum != decoded[0] {
			return status.Errorf(codes.InvalidArgument, "invalid pagination token: %v: wrong checksum", pager.pageToken)
		}
	}

	n, err := pager.c.Count(ctx)
	if err != nil {
		return errors.Trace(err)
	}
	pager.TotalSize = int32(n)

	if start > 0 && start >= n {
		return status.Errorf(codes.InvalidArgument, "invalid pagination token: start is after end: %d > %d", start, n)
	}

	c = c.Offset(start)
	if err := c.GetAll(ctx, models); err != nil {
		return errors.Trace(err)
	}

	pager.NextPageToken = ""
	end := start + int64(reflect.ValueOf(models).Elem().Len())
	if n > end {
		pager.NextPageToken, err = h.EncodeInt64([]int64{checksum, end})
		if err != nil {
			return errors.Trace(err)
		}
	}

	pager.PrevPageToken = ""
	prev := start - int64(pager.pageSize)
	if prev > 0 {
		pager.PrevPageToken, err = h.EncodeInt64([]int64{checksum, prev})
		if err != nil {
			return errors.Trace(err)
		}
	}

	return nil
}
