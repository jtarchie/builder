package builder

import (
	"path/filepath"
	"strings"

	"github.com/gosimple/slug"
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

func (d *ViewDoc) SlugPath() string {
	path := strings.Replace(
		d.filename,
		d.sourcePath,
		"",
		1,
	)

	base := filepath.Base(path)
	parts := strings.Split(filepath.Base(path), ".")
	parts[0] = slug.Make(parts[0] + " " + d.Title())
	newBase := strings.Join(parts, ".")

	path = strings.Replace(
		path,
		base,
		newBase,
		1,
	)

	path = strings.Replace(
		path,
		".md",
		".html",
		1,
	)

	return path
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
