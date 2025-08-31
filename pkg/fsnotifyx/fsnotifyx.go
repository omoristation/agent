package fsnotifyx

import (
	"context"
	"errors"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
)

const defaultTimeout = time.Minute * 3

var ErrTimeout = errors.New("fsnotifyx: timeout")

func ExitOnDeleteFile(ctx context.Context, logFunc func(format string, v ...interface{}), filePath string) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	err = watcher.Add(filepath.Dir(filePath))
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					cancel()
					return
				}
				if event.Name == filePath && event.Has(fsnotify.Remove) {
					logFunc("fsnotifyx: file %s removed, exiting...", filePath)
					cancel()
					return
				}
			case werr := <-watcher.Errors:
				logFunc("fsnotifyx: %v", werr)
			}
		}
	}()

	timeout := time.NewTimer(defaultTimeout)
	for {
		select {
		case <-timeout.C:
			return ErrTimeout
		case <-ctx.Done():
			return nil
		}
	}
}
