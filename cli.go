package builder

import (
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/fsnotify/fsnotify"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type CLI struct {
	SourcePath     string `help:"source of all files" required:"" type:"path"`
	BuildPath      string `help:"where generated content should go" required:"" type:"path"`
	LayoutFilename string `help:"layout file to render" required:"" default:"layout.html"`
	Serve          bool   `help:"serve when done building"`
}

func (c *CLI) Run() error {
	renderer := NewRender(
		filepath.Join(c.SourcePath, c.LayoutFilename),
		c.SourcePath,
		c.BuildPath,
	)

	err := renderer.Execute()
	if err != nil {
		return fmt.Errorf("could not execute render: %w", err)
	}

	if c.Serve {
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			return fmt.Errorf("could not create file watcher: %w", err)
		}
		defer watcher.Close()

		go func() {
			for event := range watcher.Events {
				matched, _ := doublestar.Match(filepath.Join(c.SourcePath, "**", "*.md"), event.Name)

				if matched {
					slog.Info("rebuilding markdown files")

					_ = renderer.Execute()
				}
			}
		}()

		err = watcher.Add(c.SourcePath)
		if err != nil {
			return fmt.Errorf("could add watching path: %w", err)
		}

		e := echo.New()
		e.Use(middleware.Logger())
		e.Static("/", c.BuildPath)

		err = e.Start(":8080")
		if err != nil {
			return fmt.Errorf("could not start serving: %w", err)
		}
	}

	return nil
}
