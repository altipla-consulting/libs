package hosting

import "context"

type Platform interface {
	Init() error
	Shutdown(ctx context.Context) error
}
