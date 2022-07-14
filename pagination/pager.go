package pagination

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"libs.altipla.consulting/database"
	"libs.altipla.consulting/errors"
)

// Deprecated: Use NewSQL instead.
type Pager struct {
	NextPageToken string
	TotalSize     int32
	PrevPageToken string

	ctrl *TokenController
	c    *database.Collection
}

// Deprecated: Use NewSQL instead.
func NewPager(c *database.Collection) *Pager {
	return &Pager{
		c:    c,
		ctrl: NewSQLToken(c, FromEmpty()),
	}
}

func (pager *Pager) SetInputs(pageToken string, pageSize int32) {
	pager.ctrl = NewSQLToken(pager.c, FromToken(pageSize, pageToken))
}

func (pager *Pager) Fetch(ctx context.Context, models interface{}) error {
	if err := pager.ctrl.Fetch(ctx, models); err != nil {
		if errors.Is(err, ErrInvalidToken) {
			return status.Errorf(codes.InvalidArgument, "%s: %v", err.Error(), pager.ctrl.token)
		}

		return errors.Trace(err)
	}

	pager.NextPageToken = pager.ctrl.NextPageToken()
	pager.PrevPageToken = pager.ctrl.PrevPageToken()
	pager.TotalSize = int32(pager.ctrl.TotalSize())

	return nil
}
