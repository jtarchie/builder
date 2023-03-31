package builder

import (
	"bytes"
	"fmt"
	htmlTemplate "html/template"
	"os"
	"path/filepath"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
	"go.abhg.dev/goldmark/frontmatter"
	"go.uber.org/zap"
)

type markdownDoc struct {
	filename  string
	sourceDir string
	buildDir  string
	logger    *zap.Logger
}

type metadataPayload struct {
	Title       string `yaml:"title"`
	Description string `yaml:"description"`
}

func NewMarkdownDoc(
	filename string,
	sourceDir string,
	buildDir string,
	logger *zap.Logger,
) *markdownDoc {
	return &markdownDoc{
		filename:  filename,
		sourceDir: sourceDir,
		buildDir:  buildDir,
		logger:    logger,
	}
}

func (m *markdownDoc) Write(
	layout *htmlTemplate.Template,
	converter goldmark.Markdown,
	templates *templates,
) error {
	template, err := templates.text(m.filename)
	if err != nil {
		return fmt.Errorf("could not get template: %w", err)
	}

	newPath := strings.Replace(
		m.filename,
		m.sourceDir,
		m.buildDir,
		1,
	)
	newPath = strings.Replace(newPath, ".md", ".html", 1)
	newDir := filepath.Dir(newPath)

	m.logger.Info("create-dir",
		zap.String("file", m.filename),
		zap.String("new-file", newPath),
		zap.String("new-dir", newDir),
	)

	err = os.MkdirAll(newDir, 0777)
	if err != nil {
		return fmt.Errorf("could not create dir (%s): %w", newDir, err)
	}

	evaluated, rendered := &bytes.Buffer{}, &bytes.Buffer{}

	err = template.Execute(evaluated, nil)
	if err != nil {
		return fmt.Errorf("could not render (%s): %w", m.filename, err)
	}

	ctx := parser.NewContext()

	err = converter.Convert(evaluated.Bytes(), rendered, parser.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("could not convert file (%s): %w", newPath, err)
	}

	file, err := os.Create(newPath)
	if err != nil {
		return fmt.Errorf("could not create file (%s): %w", newPath, err)
	}

	d := frontmatter.Get(ctx)
	if d == nil {
		return fmt.Errorf("frontmatter required (%s)", m.filename)
	}

	meta := &metadataPayload{}

	if err := d.Decode(meta); err != nil {
		return fmt.Errorf("could not decode front matter (%s): %w", m.filename, err)
	}

	err = layout.Execute(file, map[string]any{
		"Title":       meta.Title,
		"Description": meta.Description,
		"Page":        htmlTemplate.HTML(rendered.String()),
	})

	if err != nil {
		return fmt.Errorf("could not write file (%s): %w", newPath, err)
	}

	err = file.Close()
	if err != nil {
		return fmt.Errorf("could not close file (%s): %w", newPath, err)
	}

	return nil
}
