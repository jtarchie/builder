package main

import (
	"bytes"
	"fmt"
	htmlTemplate "html/template"
	"log"
	"os"
	"path/filepath"
	"strings"
	textTemplate "text/template"

	"github.com/alecthomas/kong"
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

func htmlRender(filename string) (*htmlTemplate.Template, error) {
	t, err := htmlTemplate.ParseFiles(filename)
	if err != nil {
		return nil, fmt.Errorf("could not parse HTML template (%s): %w", filename, err)
	}

	return t, nil
}

func textRender(filename string) (*textTemplate.Template, error) {
	t, err := textTemplate.ParseFiles(filename)
	if err != nil {
		return nil, fmt.Errorf("could not parse text template (%s): %w", filename, err)
	}

	return t, nil
}

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
		layoutPath := filepath.Join(c.SourcePath, c.LayoutFilename)
		layout, err := htmlRender(layoutPath)
		if err != nil {
			return fmt.Errorf("could not get layout (%s): %w", layoutPath, err)
		}

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

			markdown, err := textRender(markdownPath)
			if err != nil {
				return fmt.Errorf("could not get (%s): %w", layoutPath, err)
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

			evaluated, rendered := &bytes.Buffer{}, &bytes.Buffer{}
			err = markdown.Execute(evaluated, "test")
			if err != nil {
				return fmt.Errorf("could note render (%s): %w", markdownPath, err)
			}

			ctx := parser.NewContext()

			err = converter.Convert(evaluated.Bytes(), rendered, parser.WithContext(ctx))
			if err != nil {
				return fmt.Errorf("could not convert file (%s): %w", newPath, err)
			}

			// save to correct directory
			file, err := os.Create(newPath)
			if err != nil {
				return fmt.Errorf("could not create file (%s): %w", newPath, err)
			}

			d := frontmatter.Get(ctx)
			if d == nil {
				return fmt.Errorf("could not get front matter (%s)", markdownPath)
			}

			meta := map[string]string{}
			if err := d.Decode(&meta); err != nil {
				return fmt.Errorf("could not decode front matter (%s): %w", markdownPath, err)
			}

			err = layout.Execute(file, map[string]any{
				"Title":       meta["title"],
				"Description": meta["description"],
				"Page":        htmlTemplate.HTML(rendered.String()),
			})
			if err != nil {
				return fmt.Errorf("could not write file (%s): %w", newPath, err)
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
