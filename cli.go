package builder

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/bmatcuk/doublestar/v4"
	cp "github.com/otiai10/copy"
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
	templates := NewTemplates(logger, c.SourcePath)
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
			extension.DefinitionList,
			extension.Footnote,
			extension.Typographer,
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

	if _, err := os.Stat(assetsPath); !os.IsNotExist(err) {
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
