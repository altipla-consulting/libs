package pagination

import (
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"reflect"
	"strconv"

	"github.com/golang/protobuf/proto"

	"libs.altipla.consulting/database"
	"libs.altipla.consulting/errors"
	pb "libs.altipla.consulting/protos/pagination"
)

const DefaultPageSize = 50

type Pager struct {
	NextPageToken string
	TotalSize     int32

	c         *database.Collection
	params    []*pagerParam
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

	pager.RegisterParam("PageSize", pageSize)
}

type pagerParam struct {
	key   string
	value interface{}
}

func (pager *Pager) RegisterParam(key string, value interface{}) {
	pager.params = append(pager.params, &pagerParam{key, value})
}

func (pager *Pager) Fetch(models interface{}) error {
	checksums := []byte{}
	for _, param := range pager.params {
		checksums = append(checksums, []byte(param.key)...)
		checksums = append(checksums, []byte(fmt.Sprintf("%+v", param.value))...)
	}
	md5Checksum := md5.Sum(checksums)
	paramsChecksum := base64.StdEncoding.EncodeToString(md5Checksum[:])

	var start int64
	if pager.pageToken != "" {
		decoded, err := base64.StdEncoding.DecodeString(pager.pageToken)
		if err != nil {
			return errors.Wrapf(err, "cannot decode token")
		}
		status := new(pb.Status)
		if err := proto.Unmarshal(decoded, status); err != nil {
			return errors.Wrapf(err, "cannot unmarshal token")
		}

		if paramsChecksum != status.ParamsChecksum {
			return errors.Errorf("wrong pager status")
		}

		if status.Cursor != "" {
			start, err = strconv.ParseInt(status.Cursor, 10, 64)
			if err != nil {
				return errors.Wrapf(err, "cannot decode cursor")
			}
		}
	}

	c := pager.c.Clone().Offset(start).Limit(int64(pager.pageSize))

	if err := c.GetAll(models); err != nil {
		return errors.Trace(err)
	}

	n, err := pager.c.Count()
	if err != nil {
		return errors.Wrapf(err, "cannot count records")
	}
	pager.TotalSize = int32(n)

	end := start + int64(reflect.ValueOf(models).Elem().Len())
	if int64(pager.TotalSize) > end {
		token, err := proto.Marshal(&pb.Status{
			ParamsChecksum: paramsChecksum,
			Cursor:         fmt.Sprintf("%d", end),
		})
		if err != nil {
			return errors.Wrapf(err, "cannot marshal token")
		}
		pager.NextPageToken = base64.StdEncoding.EncodeToString(token)
	}

	return nil
}
