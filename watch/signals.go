package watch

import (
	"context"
	"fmt"
	"os"
	"os/signal"
)

func Interrupt(cancel context.CancelFunc) {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt)
	for range c {
		// Print a newline to avoid messing the output before closing
		fmt.Println()
		cancel()
		return
	}
}
