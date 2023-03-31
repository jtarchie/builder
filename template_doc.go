package builder

import (
	"path/filepath"
	"strings"
)

type TemplateDoc struct {
	*Doc
	sourcePath string
}

func (d *TemplateDoc) Path() string {
	path := strings.Replace(
		d.filename,
		d.sourcePath,
		"",
		1,
	)

	return strings.Replace(
		path,
		".md",
		".html",
		1,
	)
}

func (d *TemplateDoc) Basename() string {
	basename := filepath.Base(d.Path())

	return strings.Replace(
		basename,
		".md",
		"",
		1,
	)
}

type TemplateDocs []TemplateDoc
