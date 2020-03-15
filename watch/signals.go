package watch

import (
	"context"
	"fmt"
	"os"
	"os/signal"
)

func Interrupt(ctx context.Context, cancel context.CancelFunc) {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt)
	select {
	case <-ctx.Done():
	case <-c:
		// Print a newline to avoid messing the output before closing
		fmt.Println()
		cancel()
	}
}
