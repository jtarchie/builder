package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/bmatcuk/doublestar/v4"
	cp "github.com/otiai10/copy"
	"github.com/yuin/goldmark"
	emoji "github.com/yuin/goldmark-emoji"
	highlighting "github.com/yuin/goldmark-highlighting"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
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

	// copy public directory files over to top level
	assetsPath := filepath.Join(c.SourcePath, "public")

	logger.Info("copying",
		zap.String("build_path", c.BuildPath),
		zap.String("assets_path", assetsPath),
	)

	err = cp.Copy(assetsPath, c.BuildPath)
	if err != nil {
		return fmt.Errorf("could not copy assets contents (%s): %w", assetsPath, err)
	}

	// go through each markdown
	{
		logger := logger.Named("markdown")
		pattern := filepath.Join(c.SourcePath, "**", "*.md")

		logger.Info("glob",
			zap.String("pattern", pattern),
		)

		matches, err := doublestar.FilepathGlob(pattern)
		if err != nil {
			return fmt.Errorf("could not glob markdown files: %w", err)
		}

		converter := goldmark.New(
			goldmark.WithExtensions(
				extension.GFM,
				emoji.Emoji,
				&frontmatter.Extender{},
				&mermaid.Extender{},
				highlighting.Highlighting,
			),
			goldmark.WithParserOptions(
				parser.WithAutoHeadingID(),
			),
		)

		for _, markdownPath := range matches {
			// render to HTML within layout
			logger.Info("reading",
				zap.String("file", markdownPath),
			)

			contents, err := os.ReadFile(markdownPath)
			if err != nil {
				return fmt.Errorf("could not read file (%s): %w", markdownPath, err)
			}

			newPath := strings.Replace(
				markdownPath,
				c.SourcePath,
				c.BuildPath,
				1,
			)
			newPath = strings.Replace(
				newPath,
				".md",
				".html",
				1,
			)
			newDir := filepath.Dir(newPath)

			logger.Info("create-dir",
				zap.String("file", markdownPath),
				zap.String("new-file", newPath),
				zap.String("new-dir", newDir),
			)

			err = os.MkdirAll(newDir, 0777)
			if err != nil {
				return fmt.Errorf("could not create dir (%s): %w", newDir, err)
			}

			buffer := &bytes.Buffer{}
			err = converter.Convert(contents, buffer)
			if err != nil {
				return fmt.Errorf("could not convert file (%s): %w", newPath, err)
			}

			// save to correct directory
			file, err := os.Create(newPath)
			if err != nil {
				return fmt.Errorf("could not create file (%s): %w", newPath, err)
			}

			err = file.Close()
			if err != nil {
				return fmt.Errorf("could not close file (%s): %w", newPath, err)
			}
		}
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
