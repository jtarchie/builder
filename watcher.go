package builder

import (
	"fmt"

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

func (w *Watcher) Execute(fn func(string) error) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("could not create file watcher: %w", err)
	}
	defer watcher.Close()

	err = watcher.Add(w.sourceDir)
	if err != nil {
		return fmt.Errorf("could add watching path: %w", err)
	}

	for event := range watcher.Events {
		err := fn(event.Name)
		if err != nil {
			return fmt.Errorf("could not execute fn in watcher: %w", err)
		}
	}

	return nil
}
