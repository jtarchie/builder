package builder

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"

	"github.com/Masterminds/sprig/v3"
	"github.com/gosimple/slug"
	cp "github.com/otiai10/copy"
	"github.com/tdewolff/minify"
	mHTML "github.com/tdewolff/minify/html"
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
	converter  goldmark.Markdown
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
		converter: goldmark.New(
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
		),
	}
}

func (r *Render) Execute(pattern string) error {
	err := r.copyAssets()
	if err != nil {
		return fmt.Errorf("copying assets issue: %w", err)
	}

	docs, err := NewDocs(r.sourcePath, pattern, 0, false)
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
		"iterDocs": func(path string, limit int) (Docs, error) {
			pattern := filepath.Join(r.sourcePath, path, "*.md")
			docs, err := NewDocs(r.sourcePath, pattern, limit, true)
			if err != nil {
				return nil, fmt.Errorf("could not load docs: %w", err)
			}

			return docs, nil
		},
	}

	for _, doc := range docs {
		err := r.renderMarkdown(doc, funcMap, layout)
		if err != nil {
			return fmt.Errorf("rendering template issue: %w", err)
		}
	}

	return nil
}

func (r *Render) copyAssets() error {
	err := os.RemoveAll(r.buildPath)
	if err != nil {
		return fmt.Errorf("could not remove build path (%s): %w", r.buildPath, err)
	}

	err = os.MkdirAll(r.buildPath, os.ModePerm)
	if err != nil {
		return fmt.Errorf("could not create build path (%s): %w", r.buildPath, err)
	}

	assetsPath := filepath.Join(r.sourcePath, "public")

	if _, err := os.Stat(assetsPath); !os.IsNotExist(err) {
		err = cp.Copy(assetsPath, r.buildPath)
		if err != nil {
			return fmt.Errorf("could not copy assets contents (%s): %w", assetsPath, err)
		}
	}

	return nil
}

func (r *Render) renderMarkdown(doc *Doc, funcMap template.FuncMap, layout *template.Template) error {
	match := doc.Filename()

	if doc.Title() == "" {
		return fmt.Errorf("could not determine title (%s)", match)
	}

	markdown, err := template.
		New(match).
		Funcs(funcMap).
		Funcs(sprig.FuncMap()).
		Parse(doc.Contents())
	if err != nil {
		return fmt.Errorf("could not parse markdown template (%s): %w", r.layoutPath, err)
	}

	layoutWriter, markdownWriter, renderedWriter := &bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{}

	err = markdown.Execute(markdownWriter, nil)
	if err != nil {
		return fmt.Errorf("could not render markdown template (%s): %w", match, err)
	}

	err = r.converter.Convert(markdownWriter.Bytes(), renderedWriter)
	if err != nil {
		return fmt.Errorf("could not convert file (%s): %w", match, err)
	}

	err = layout.Execute(layoutWriter, map[string]any{
		"Doc":          doc,
		"RenderedPage": template.HTML(renderedWriter.String()),
	})
	if err != nil {
		return fmt.Errorf("could not render layout template (%s): %w", match, err)
	}

	withoutSlugFilename := strings.Replace(doc.Filename(), r.sourcePath, r.buildPath, 1)
	withoutSlugFilename = strings.Replace(withoutSlugFilename, ".md", ".html", 1)
	filenames := []string{withoutSlugFilename}

	if !strings.Contains(withoutSlugFilename, "index.html") {
		withSlugFilename := strings.Replace(withoutSlugFilename, ".html", "-"+slug.Make(doc.Title())+".html", 1)
		filenames = append(filenames, withSlugFilename)
	}

	err = writeHTMLFiles(filenames, layoutWriter.String())
	if err != nil {
		return fmt.Errorf("could write new template: %w", err)
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

func writeHTMLFiles(filenames []string, contents string) error {
	dirPath := filepath.Dir(filenames[0])

	err := os.MkdirAll(dirPath, os.ModePerm)
	if err != nil {
		return fmt.Errorf("could not create path (%s): %w", dirPath, err)
	}

	writer := &strings.Builder{}

	err = mHTML.Minify(&minify.M{}, writer, strings.NewReader(contents), nil)
	if err != nil {
		return fmt.Errorf("could not minify: %w", err)
	}

	for _, filename := range filenames {
		err = os.WriteFile(filename, []byte(writer.String()), os.ModePerm)
		if err != nil {
			return fmt.Errorf("could not write path (%s): %w", filename, err)
		}
	}

	return nil
}
