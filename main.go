package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	fm "github.com/adrg/frontmatter"
	"github.com/alecthomas/kong"
	"github.com/bmatcuk/doublestar/v4"
	cp "github.com/otiai10/copy"
	"github.com/samber/lo"
	"github.com/yuin/goldmark"
	emoji "github.com/yuin/goldmark-emoji"
	highlighting "github.com/yuin/goldmark-highlighting"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"go.abhg.dev/goldmark/frontmatter"
	"go.abhg.dev/goldmark/mermaid"
	"go.uber.org/zap"
)

type CLI struct {
	SourcePath     string `help:"source of all files" required:"" type:"path"`
	BuildPath      string `help:"where generated content should go" required:"" type:"path"`
	LayoutFilename string `help:"layout file to render" required:"" default:"layout.html"`
}

func (c *CLI) Run(logger *zap.Logger) error {
	// rm build dir
	err := c.buildSetup(logger)
	if err != nil {
		return fmt.Errorf("build setup failed: %w", err)
	}

	// copy public directory files over to top level
	err = c.assetSetup(logger)
	if err != nil {
		return fmt.Errorf("asset setup failed: %w", err)
	}

	// go through each markdown
	templateLogger := logger.Named("template.func")

	templates := templates{
		FuncMap: map[string]any{
			"iterDocs": func(path string, limit int) []Doc {
				pattern := filepath.Join(c.SourcePath, path, "**", "*.md")

				matches, err := doublestar.FilepathGlob(pattern)
				if err != nil {
					templateLogger.Fatal("glob",
						zap.String("pattern", pattern),
						zap.Error(err),
					)
				}

				matches = lo.Filter(matches, func(path string, _ int) bool {
					return !strings.HasSuffix(path, "index.md")
				})

				var docs []Doc
				if len(matches) > limit {
					matches = matches[:limit]
				}

				for _, match := range matches {
					contents, err := os.ReadFile(match)
					if err != nil {
						templateLogger.Fatal("read",
							zap.String("match", match),
							zap.Error(err),
						)
					}
					metadata := map[string]string{}

					_, err = fm.Parse(bytes.NewReader(contents), &metadata)
					if err != nil {
						templateLogger.Fatal("metadata",
							zap.String("match", match),
							zap.Error(err),
						)
					}

					docs = append(docs, Doc{
						Title: metadata["title"],
						Path: strings.Replace(
							strings.Replace(match, c.SourcePath, "", 1),
							".md",
							".html",
							1,
						),
						BaseName: filepath.Base(match),
					})
				}

				return docs
			},
		},
	}

	layoutPath := filepath.Join(c.SourcePath, c.LayoutFilename)

	layout, err := templates.html(layoutPath)
	if err != nil {
		return fmt.Errorf("could not get layout (%s): %w", layoutPath, err)
	}

	logger = logger.Named("markdown")
	pattern := filepath.Join(c.SourcePath, "**", "*.md")

	logger.Info("glob",
		zap.String("pattern", pattern),
	)

	matches, err := doublestar.FilepathGlob(pattern)
	if err != nil {
		return fmt.Errorf("could not glob markdown files: %w", err)
	}

	converter := goldmark.New(
		goldmark.WithRendererOptions(
			html.WithXHTML(),
			html.WithUnsafe(),
		),
		goldmark.WithExtensions(
			&frontmatter.Extender{},
			extension.GFM,
			emoji.Emoji,
			&mermaid.Extender{},
			highlighting.Highlighting,
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
			parser.WithAttribute(),
		),
	)

	for _, markdownPath := range matches {
		// render to HTML within layout
		logger.Info("reading",
			zap.String("file", markdownPath),
		)

		markdown := NewMarkdownDoc(
			markdownPath,
			c.SourcePath,
			c.BuildPath,
			logger,
		)

		// save to correct directory
		err := markdown.Write(
			layout,
			converter,
			templates,
		)
		if err != nil {
			return fmt.Errorf("could not render markdown: %w", err)
		}
	}

	return nil
}

func (c *CLI) assetSetup(logger *zap.Logger) error {
	assetsPath := filepath.Join(c.SourcePath, "public")

	logger.Info("copying",
		zap.String("build_path", c.BuildPath),
		zap.String("assets_path", assetsPath),
	)

	if _, err := os.Stat("/path/to/whatever"); !os.IsNotExist(err) {
		err = cp.Copy(assetsPath, c.BuildPath)
		if err != nil {
			return fmt.Errorf("could not copy assets contents (%s): %w", assetsPath, err)
		}
	} else {
		logger.Info("copying.skipping",
			zap.String("build_path", c.BuildPath),
			zap.String("assets_path", assetsPath),
		)
	}

	return nil
}

func (c *CLI) buildSetup(logger *zap.Logger) error {
	logger.Info("removing",
		zap.String("build_path", c.BuildPath),
	)

	err := os.RemoveAll(c.BuildPath)
	if err != nil {
		return fmt.Errorf("could not remove build path (%s): %w", c.BuildPath, err)
	}

	logger.Info("removing",
		zap.String("build_path", c.BuildPath),
	)

	err = os.MkdirAll(c.BuildPath, 0777)
	if err != nil {
		return fmt.Errorf("could not create build path (%s): %w", c.BuildPath, err)
	}

	return nil
}

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("could not start logger: %s", err)
	}

	cli := &CLI{}
	ctx := kong.Parse(cli)
	// Call the Run() method of the selected parsed command.
	err = ctx.Run(logger)
	ctx.FatalIfErrorf(err)
}
