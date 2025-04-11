package builder

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

type Watcher struct {
	sourceDir string
}

func NewWatcher(
	sourceDir string,
) *Watcher {
	return &Watcher{
		sourceDir: sourceDir,
	}
}

func (w *Watcher) Execute(watchFn func(string) error) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("could not create file watcher: %w", err)
	}
	defer func() { _ = watcher.Close() }()

	err = filepath.Walk(w.sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && info.Name()[0] != '.' {
			err = watcher.Add(path)
			if err != nil {
				return fmt.Errorf("could not add watching path %s: %w", path, err)
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("could not walk through source directory: %w", err)
	}

	for event := range watcher.Events {
		err := watchFn(event.Name)
		if err != nil {
			return fmt.Errorf("could not execute fn in watcher: %w", err)
		}
	}

	return nil
}
