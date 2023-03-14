package pagination

import (
	"context"

	"github.com/altipla-consulting/errors"

	"libs.altipla.consulting/database"
	"libs.altipla.consulting/rdb"
)

type storageAdapter interface {
	checksum(limit int32) uint32
	fetch(ctx context.Context, models interface{}, start int64, pageSize int32) (int64, error)
}

type rdbStorage struct {
	q        *rdb.Query
	includes []rdb.IncludeOption
}

func newRDBStorage(q *rdb.Query) storageAdapter {
	return &rdbStorage{q: q}
}

func (storage *rdbStorage) checksum(limit int32) uint32 {
	return storage.q.Clone().Limit(int64(limit)).Checksum()
}

func (storage *rdbStorage) fetch(ctx context.Context, models interface{}, start int64, pageSize int32) (int64, error) {
	q := storage.q.Clone().Limit(int64(pageSize)).Offset(start)
	if err := q.GetAll(ctx, models, storage.includes...); err != nil {
		return 0, errors.Trace(err)
	}

	return q.Stats().TotalResults, nil
}

type sqlStorage struct {
	q *database.Collection
}

func newSQLStorage(q *database.Collection) storageAdapter {
	return &sqlStorage{q: q}
}

func (storage *sqlStorage) checksum(limit int32) uint32 {
	return storage.q.Clone().Limit(int64(limit)).Checksum()
}

func (storage *sqlStorage) fetch(ctx context.Context, models interface{}, start int64, pageSize int32) (int64, error) {
	if err := storage.q.Clone().Limit(int64(pageSize)).Offset(start).GetAll(ctx, models); err != nil {
		return 0, errors.Trace(err)
	}

	total, err := storage.q.Clone().Count(ctx)
	if err != nil {
		return 0, errors.Trace(err)
	}

	return total, nil
}
