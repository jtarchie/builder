package builder

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/gosimple/slug"
	cp "github.com/otiai10/copy"
	"github.com/samber/lo"
	"github.com/yuin/goldmark"
	emoji "github.com/yuin/goldmark-emoji"
	highlighting "github.com/yuin/goldmark-highlighting"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"go.abhg.dev/goldmark/mermaid"
)

type Render struct {
	layoutPath string
	sourcePath string
	buildPath  string
}

func NewRender(
	layoutPath string,
	sourcePath string,
	buildPath string,
) *Render {
	return &Render{
		layoutPath: layoutPath,
		sourcePath: sourcePath,
		buildPath:  buildPath,
	}
}

func (r *Render) Execute() error {
	// create build directory
	err := os.RemoveAll(r.buildPath)
	if err != nil {
		return fmt.Errorf("could not remove build path (%s): %w", r.buildPath, err)
	}

	err = os.MkdirAll(r.buildPath, 0777)
	if err != nil {
		return fmt.Errorf("could not create build path (%s): %w", r.buildPath, err)
	}

	// copy over assets
	assetsPath := filepath.Join(r.sourcePath, "public")

	if _, err := os.Stat(assetsPath); !os.IsNotExist(err) {
		err = cp.Copy(assetsPath, r.buildPath)
		if err != nil {
			return fmt.Errorf("could not copy assets contents (%s): %w", assetsPath, err)
		}
	}

	// foreach markdown file
	pattern := filepath.Join(r.sourcePath, "**", "*.md")

	matches, err := doublestar.FilepathGlob(pattern)
	if err != nil {
		return fmt.Errorf("could not glob markdown files: %w", err)
	}

	contents, err := readFile(r.layoutPath)
	if err != nil {
		return fmt.Errorf("could not read layout: %w", err)
	}

	layout, err := template.New(r.layoutPath).Parse(contents)
	if err != nil {
		return fmt.Errorf("could not parse layout template (%s): %w", r.layoutPath, err)
	}

	funcMap := template.FuncMap{
		"iterDocs": func(path string, limit int) ([]ViewDoc, error) {
			pattern := filepath.Join(r.sourcePath, path, "*.md")
			docs, err := NewDocs(pattern, limit)
			if err != nil {
				return nil, fmt.Errorf("could not load docs: %w", err)
			}

			return lo.Map(docs, func(doc *Doc, _ int) ViewDoc {
				return ViewDoc{
					Doc:        doc,
					sourcePath: r.sourcePath,
				}
			}), nil
		},
	}

	for _, match := range matches {
		//   get filename
		doc, err := NewDoc(match)
		if err != nil {
			return fmt.Errorf("could not read markdown doc (%s): %w", match, err)
		}

		if doc.Title() == "" {
			return fmt.Errorf("could not determine title (%s)", match)
		}

		markdown, err := template.New(match).Funcs(funcMap).Parse(doc.Contents())
		if err != nil {
			return fmt.Errorf("could not parse markdown template (%s): %w", r.layoutPath, err)
		}

		layoutWriter, markdownWriter, renderedWriter := &bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{}

		err = markdown.Execute(markdownWriter, nil)
		if err != nil {
			return fmt.Errorf("could not render markdown template (%s): %w", match, err)
		}

		err = converter.Convert(markdownWriter.Bytes(), renderedWriter)
		if err != nil {
			return fmt.Errorf("could not convert file (%s): %w", match, err)
		}

		err = layout.Execute(layoutWriter, map[string]any{
			"Title":       doc.Title(),
			"Description": doc.Description(),
			"Page":        template.HTML(renderedWriter.String()),
		})
		if err != nil {
			return fmt.Errorf("could not render layout template (%s): %w", match, err)
		}

		newFilename := strings.Replace(match, r.sourcePath, r.buildPath, 1)
		newFilename = strings.Replace(newFilename, ".md", ".html", 1)

		err = writeFile(newFilename, layoutWriter.String())
		if err != nil {
			return fmt.Errorf("could write new template (%s): %w", newFilename, err)
		}

		if !strings.Contains(newFilename, "index.html") {
			newFilename = strings.Replace(newFilename, ".html", "-"+slug.Make(doc.Title())+".html", 1)

			err = writeFile(newFilename, layoutWriter.String())
			if err != nil {
				return fmt.Errorf("could write new template (%s): %w", newFilename, err)
			}
		}
	}

	return nil
}

func readFile(filename string) (string, error) {
	contents, err := os.ReadFile(filename)
	if err != nil {
		return "", fmt.Errorf("could not read file (%s): %w", filename, err)
	}

	return string(contents), nil
}

func writeFile(filename, contents string) error {
	dirPath := filepath.Dir(filename)

	err := os.MkdirAll(dirPath, 0777)
	if err != nil {
		return fmt.Errorf("could not create path (%s): %w", dirPath, err)
	}

	err = os.WriteFile(filename, []byte(contents), 0777)
	if err != nil {
		return fmt.Errorf("could not write path (%s): %w", filename, err)
	}

	return nil
}

var converter = goldmark.New(
	goldmark.WithRendererOptions(
		html.WithXHTML(),
		html.WithUnsafe(),
	),
	goldmark.WithExtensions(
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
