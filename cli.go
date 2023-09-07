package builder

import (
	"fmt"
	"log"
	"log/slog"
	"path/filepath"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/jtarchie/builder/indexers"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type CLI struct {
	BuildPath      string `help:"where generated content should go" required:"" type:"path"`
	Index          bool   `help:"index the documents for fusejs usage"`
	LayoutFilename string `help:"layout file to render" required:"" default:"layout.html"`
	Serve          bool   `help:"serve when done building"`
	SourcePath     string `help:"source of all files" required:"" type:"path"`
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

	if c.Index {
		indexer := indexers.NewFuseJS(c.BuildPath)

		err = indexer.Execute()
		if err != nil {
			return fmt.Errorf("could not execute indexer: %w", err)
		}
	}

	if c.Serve {
		watcher := NewWatcher(c.SourcePath)

		go func() {
			err := watcher.Execute(func(filename string) error {
				glob := filepath.Join(c.SourcePath, "**", "*.md")
				matched, _ := doublestar.Match(glob, filename)

				if matched {
					slog.Info("rebuilding markdown files")

					err := renderer.Execute()
					if err != nil {
						slog.Error("could not rebuild markdown files", slog.String("error", err.Error()))
					}
				}

				return nil
			})
			if err != nil {
				log.Fatalf("could not run watcher: %s", err)
			}
		}()

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
