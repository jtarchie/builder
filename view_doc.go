package builder

import (
	"path/filepath"
	"strings"
)

type ViewDoc struct {
	*Doc
	sourcePath string
}

func (d *ViewDoc) Path() string {
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

func (d *ViewDoc) Basename() string {
	basename := filepath.Base(d.Path())

	return strings.Replace(
		basename,
		".html",
		"",
		1,
	)
}
