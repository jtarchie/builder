package builder

import (
	"bytes"
	"fmt"
	htmlTemplate "html/template"
	"os"
	"path/filepath"
	"sort"
	"strings"
	textTemplate "text/template"

	"github.com/adrg/frontmatter"
	"github.com/bmatcuk/doublestar/v4"
	"github.com/samber/lo"
	"go.uber.org/zap"
)

type Doc struct {
	Title    string
	Path     string
	BaseName string
}

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

func (f *templates) text(filename string) (*textTemplate.Template, error) {
	contents, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("could not read file template (%s): %w", filename, err)
	}

	t, err := textTemplate.New(filename).Funcs(f.FuncMap).Parse(string(contents))
	if err != nil {
		return nil, fmt.Errorf("could not parse text template (%s): %w", filename, err)
	}

	return t, nil
}

func NewTemplates(
	logger *zap.Logger,
	sourcePath string,
) *templates {
	templateLogger := logger.Named("template.func")

	return &templates{
		FuncMap: map[string]any{
			"iterDocs": func(path string, limit int) []Doc {
				pattern := filepath.Join(sourcePath, path, "**", "*.md")

				matches, err := doublestar.FilepathGlob(pattern)
				if err != nil {
					templateLogger.Fatal("glob",
						zap.String("pattern", pattern),
						zap.Error(err),
					)
				}

				matches = lo.Filter(matches, func(path string, _ int) bool {
					return !strings.HasSuffix(path, "index.md")
				})

				sort.Strings(matches)
				sort.Sort(sort.Reverse(sort.StringSlice(matches)))

				var docs []Doc
				if len(matches) > limit && limit > 0 {
					matches = matches[:limit]
				}

				for _, match := range matches {
					contents, err := os.ReadFile(match)
					if err != nil {
						templateLogger.Fatal("read",
							zap.String("match", match),
							zap.Error(err),
						)
					}
					metadata := map[string]string{}

					_, err = frontmatter.Parse(bytes.NewReader(contents), &metadata)
					if err != nil {
						templateLogger.Fatal("metadata",
							zap.String("match", match),
							zap.Error(err),
						)
					}

					docs = append(docs, Doc{
						Title: metadata["title"],
						Path: changeExtension(
							strings.Replace(match, sourcePath, "", 1),
						),
						BaseName: changeExtension(filepath.Base(match)),
					})
				}

				return docs
			},
		},
	}
}

func changeExtension(filename string) string {
	return strings.Replace(filename, ".md", ".html", 1)
}
