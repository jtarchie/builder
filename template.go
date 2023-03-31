package builder

import (
	"fmt"
	htmlTemplate "html/template"
	"os"
	"path/filepath"
	"regexp"
	textTemplate "text/template"
)

type templates struct {
	textTemplate.FuncMap
}

func (f *templates) html(filename string) (*htmlTemplate.Template, error) {
	contents, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("could not read file template (%s): %w", filename, err)
	}

	t, err := htmlTemplate.New(filename).Funcs(f.FuncMap).Parse(string(contents))
	if err != nil {
		return nil, fmt.Errorf("could not parse HTML template (%s): %w", filename, err)
	}

	return t, nil
}

func (f *templates) text(doc *Doc) (*textTemplate.Template, error) {
	t, err := textTemplate.New(doc.Filename()).Funcs(f.FuncMap).Parse(doc.Contents())
	if err != nil {
		return nil, fmt.Errorf("could not parse text template (%s): %w", doc.Filename(), err)
	}

	return t, nil
}

func NewTemplates(
	sourcePath string,
) *templates {
	return &templates{
		FuncMap: map[string]any{
			"iterDocs": func(path string, limit int) (TemplateDocs, error) {
				pattern := filepath.Join(sourcePath, path, "**", "*.md")

				docs, err := NewDocs(pattern, regexp.MustCompile(`index\.md`))
				if err != nil {
					return nil, fmt.Errorf("could not load docs: %w", err)
				}

				if len(docs) > limit && limit > 0 {
					docs = docs[:limit]
				}

				templateDocs := TemplateDocs{}
				for _, doc := range docs {
					templateDocs = append(templateDocs, TemplateDoc{
						Doc:        doc,
						sourcePath: sourcePath,
					})
				}

				return templateDocs, nil
			},
		},
	}
}
