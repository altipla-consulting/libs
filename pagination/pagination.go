package pagination

import (
	"encoding/base64"
	"reflect"

	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"libs.altipla.consulting/database"
	"libs.altipla.consulting/errors"
	pb "libs.altipla.consulting/pagination/internal/model"
)

const DefaultPageSize = 50

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
	if pager.pageSize <= 0 || pager.pageSize > DefaultPageSize {
		pager.pageSize = DefaultPageSize
	}
}

func (pager *Pager) Fetch(models interface{}) error {
	c := pager.c.Clone().Limit(int64(pager.pageSize))
	checksum := c.Checksum()

	var start int64
	if pager.pageToken != "" {
		decoded, err := base64.RawURLEncoding.DecodeString(pager.pageToken)
		if err != nil {
			return status.Errorf(codes.InvalidArgument, "invalid pagination token: %v: %v", pager.pageToken, err)
		}
		in := new(pb.Status)
		if err := proto.Unmarshal(decoded, in); err != nil {
			return status.Errorf(codes.InvalidArgument, "invalid pagination token: %v: %v", pager.pageToken, err)
		}

		start = in.Cursor

		if checksum != in.Checksum {
			return status.Errorf(codes.InvalidArgument, "invalid pagination token: %v: wrong checksum", pager.pageToken)
		}
	}

	c = c.Offset(start)
	if err := c.GetAll(models); err != nil {
		return errors.Trace(err)
	}

	n, err := pager.c.Count()
	if err != nil {
		return errors.Trace(err)
	}
	pager.TotalSize = int32(n)

	end := start + int64(reflect.ValueOf(models).Elem().Len())
	if n > end {
		token, err := proto.Marshal(&pb.Status{
			Checksum: checksum,
			Cursor:   end,
		})
		if err != nil {
			return errors.Trace(err)
		}
		pager.NextPageToken = base64.RawURLEncoding.EncodeToString(token)
	}

	if start > 0 {
		token, err := proto.Marshal(&pb.Status{
			Checksum: checksum,
			Cursor:   start - int64(pager.pageSize),
		})
		if err != nil {
			return errors.Trace(err)
		}
		pager.PrevPageToken = base64.RawURLEncoding.EncodeToString(token)
	}

	return nil
}
