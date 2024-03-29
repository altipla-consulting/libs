package watch

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"time"

	"github.com/altipla-consulting/errors"
	"github.com/fsnotify/fsnotify"
)

func Recursive(ctx context.Context, changes chan string, folders ...string) error {
	var paths []string
	walkFn := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return errors.Trace(err)
		}
		if !info.IsDir() {
			return nil
		}

		paths = append(paths, path)

		return nil
	}
	for _, folder := range folders {
		if err := filepath.Walk(folder, walkFn); err != nil {
			return errors.Trace(err)
		}
	}

	return errors.Trace(Files(ctx, changes, paths...))
}

func Files(ctx context.Context, changes chan string, paths ...string) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return errors.Trace(err)
	}

	for _, path := range paths {
		if err := watcher.Add(path); err != nil {
			return errors.Trace(err)
		}
	}

	var timer *time.Timer
	var pending []string
	for {
		var timerCh <-chan time.Time
		if timer != nil {
			timerCh = timer.C
		}

		select {
		case <-ctx.Done():
			return nil

		case err := <-watcher.Errors:
			return errors.Trace(err)

		case ev := <-watcher.Events:
			if ev.Op == fsnotify.Create {
				info, err := os.Stat(ev.Name)
				if err != nil {
					return errors.Trace(err)
				}
				if info.IsDir() {
					if err := watcher.Add(ev.Name); err != nil {
						return errors.Trace(err)
					}
				}
			}

			if timer == nil {
				timer = time.NewTimer(100 * time.Millisecond)
			}
			pending = append(pending, ev.Name)

		case <-timerCh:
			slices.Sort(pending)
			for _, change := range slices.Compact(pending) {
				changes <- change
			}
			pending = nil
			timer = nil
		}
	}
}
