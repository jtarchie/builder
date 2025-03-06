package builder

import (
	"fmt"
	"log"
	"log/slog"
	"path/filepath"

	"errors"
	"github.com/bmatcuk/doublestar/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type CLI struct {
	AssetsPath     string `help:"path to static assets (default with be source-path/public)"`
	BaseURL        string `help:"the URL which the contents will be served from, this is only used for generating feeds"`
	BuildPath      string `help:"where generated content should go"                                                      required:""                  type:"path"`
	FeedGlob       string `help:"glob patterns for documents to feature in feeds"`
	LayoutFilename string `default:"layout.html"                                                                         help:"layout file to render" required:""`
	Serve          bool   `help:"serve when done building"`
	SourcePath     string `help:"source of all files"                                                                    required:""                  type:"path"`
}

func (c *CLI) Run() error {
	if c.AssetsPath == "" {
		c.AssetsPath = filepath.Join(c.SourcePath, "public")
	}

	layoutPath := filepath.Join(c.SourcePath, c.LayoutFilename)
	if filepath.IsAbs(c.LayoutFilename) {
		layoutPath = c.LayoutFilename
	}

	renderer := NewRender(
		layoutPath,
		c.SourcePath,
		c.AssetsPath,
		c.BuildPath,
		c.BaseURL,
	)

	markdownGlob := filepath.Join(c.SourcePath, "**", "*.md")

	if c.FeedGlob == "" {
		c.FeedGlob = markdownGlob
	} else {
		c.FeedGlob = filepath.Join(c.SourcePath, c.FeedGlob)
	}

	err := renderer.Execute(
		markdownGlob,
		c.FeedGlob,
	)
	if err != nil {
		return fmt.Errorf("could not execute render: %w", err)
	}

	if c.Serve {
		slog.Info("starting server", "sourcePath", c.SourcePath, "buildPath", c.BuildPath)

		watcher := NewWatcher(c.SourcePath)

		go c.startWatcher(watcher, renderer, markdownGlob, c.FeedGlob)

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

func (c *CLI) startWatcher(
	watcher *Watcher,
	renderer *Render,
	markdownGlob string,
	feedGlob string,
) {
	allGlob := filepath.Join(c.SourcePath, "**", "{*.md,*.html,*.js,*.css}")
	slog.Info("globbing all files", slog.String("glob", allGlob))

	err := watcher.Execute(func(filename string) error {
		slog.Info("file changed", slog.String("filename", filename))

		matched, err := doublestar.Match(allGlob, filename)
		if errors.Is(err, doublestar.ErrBadPattern) {
			return fmt.Errorf("could not use glob pattern: %w", err)
		}

		if matched {
			slog.Info("rebuilding markdown files", slog.String("filename", filename))

			err := renderer.Execute(
				markdownGlob,
				feedGlob,
			)
			if err != nil {
				slog.Error("could not rebuild markdown files", slog.String("error", err.Error()))
			}
		}

		return nil
	})
	if err != nil {
		log.Fatalf("could not run watcher: %s", err)
	}
}
