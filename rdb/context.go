package rdb

import (
	"context"
)

type contextKey int

const (
	keySession = contextKey(iota)
)

func SessionFromContext(ctx context.Context) *Session {
	sess, ok := ctx.Value(keySession).(*Session)
	if !ok {
		return nil
	}
	return sess
}
