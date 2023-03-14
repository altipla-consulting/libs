package database

import (
	"context"

	"github.com/altipla-consulting/errors"
)

type HookFn func(ctx context.Context, instance Model) error

type hooker struct {
	afterPut []HookFn
}

func (h *hooker) runAfterPut(ctx context.Context, instance Model) error {
	for _, fn := range h.afterPut {
		if err := fn(ctx, instance); err != nil {
			return errors.Trace(err)
		}
	}
	return nil
}

func WithAfterPut(fn HookFn) CollectionOption {
	return func(c *Collection) {
		c.h.afterPut = append(c.h.afterPut, fn)
	}
}
