package builder

import (
	"bytes"
	"fmt"
	htmlTemplate "html/template"
	"os"
	"path/filepath"
	"strings"

	"github.com/yuin/goldmark"
	"go.uber.org/zap"
)

type markdownDoc struct {
	filename  string
	sourceDir string
	buildDir  string
	logger    *zap.Logger
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
	doc, err := NewDoc(m.filename)
	if err != nil {
		return fmt.Errorf("could not get markdown file (%s): %w", m.filename, err)
	}

	template, err := templates.text(doc)
	if err != nil {
		return fmt.Errorf("could not init template: %w", err)
	}

	newPath := strings.Replace(
		m.filename,
		m.sourceDir,
		m.buildDir,
		1,
	)
	newPath = strings.Replace(newPath, ".md", ".html", 1)
	newDir := filepath.Dir(newPath)

	err = os.MkdirAll(newDir, 0777)
	if err != nil {
		return fmt.Errorf("could not create dir (%s): %w", newDir, err)
	}

	evaluated, rendered := &bytes.Buffer{}, &bytes.Buffer{}

	err = template.Execute(evaluated, nil)
	if err != nil {
		return fmt.Errorf("could not render (%s): %w", m.filename, err)
	}

	err = converter.Convert(evaluated.Bytes(), rendered)
	if err != nil {
		return fmt.Errorf("could not convert file (%s): %w", newPath, err)
	}

	file, err := os.Create(newPath)
	if err != nil {
		return fmt.Errorf("could not create file (%s): %w", newPath, err)
	}

	if doc.Title() == "" {
		return fmt.Errorf("document has no title (%s)", m.filename)
	}

	err = layout.Execute(file, map[string]any{
		"Title":       doc.Title(),
		"Description": doc.Description(),
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
