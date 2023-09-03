package builder

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/bmatcuk/doublestar/v4"
	"github.com/dop251/goja"
	"github.com/microcosm-cc/bluemonday"
)

type Indexer struct {
	buildPath string
	vm        *goja.Runtime
}

func NewIndexer(
	buildPath string,
) *Indexer {
	vm := goja.New()
	vm.SetFieldNameMapper(goja.TagFieldNameMapper("json", true))

	return &Indexer{
		buildPath: buildPath,
		vm:        vm,
	}
}

func (i *Indexer) Execute() error {
	type docPayload struct {
		RelativePath string `json:"id"`
		Title        string `json:"title"`
		Contents     string `json:"contents"`
	}

	matches, err := doublestar.FilepathGlob(filepath.Join(i.buildPath, "**", "*.html"))
	if err != nil {
		return fmt.Errorf("could not find HTML files for indexing: %w", err)
	}

	documents := []docPayload{}

	policy := bluemonday.StrictPolicy()

	for _, filename := range matches {
		contents, err := os.ReadFile(filename)
		if err != nil {
			return fmt.Errorf("could not read file for indexing (%q): %w", filename, err)
		}

		doc, err := goquery.NewDocumentFromReader(bytes.NewReader(contents))
		if err != nil {
			return fmt.Errorf("could not parse file for HTML (%q): %w", filename, err)
		}

		documents = append(documents, docPayload{
			RelativePath: strings.Replace(filename, i.buildPath, "", 1),
			Title:        doc.Find("title").First().Text(),
			Contents:     policy.Sanitize(string(contents)),
		})
	}

	err = i.vm.Set("documents", documents)
	if err != nil {
		return fmt.Errorf("could not create documents for index: %w", err)
	}

	val, err := i.vm.RunString(searchBuildJS)
	if err != nil {
		return fmt.Errorf("could not build index: %w", err)
	}

	if json, ok := val.Export().(string); ok {
		indexFilename := filepath.Join(i.buildPath, "index.json")

		err = os.WriteFile(indexFilename, []byte(json), os.ModePerm)
		if err != nil {
			return fmt.Errorf("could not write index file (%q): %w", indexFilename, err)
		}
	}

	return nil
}
