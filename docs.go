package builder

import (
	"fmt"
	"sort"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/samber/lo"
	"github.com/samber/lo/mutable"
)

type Docs []*Doc

func NewDocs(
	sourcePath string,
	pattern string,
	limit int,
	filterIndexes bool,
) (Docs, error) {
	matches, err := doublestar.FilepathGlob(pattern)
	if err != nil {
		return nil, fmt.Errorf("could not matches (%q): %w", pattern, err)
	}

	if filterIndexes {
		matches = lo.Filter(matches, func(match string, _ int) bool {
			return !strings.HasSuffix(match, "index.md")
		})
	}

	sort.Strings(matches)
	mutable.Reverse(matches)

	if limit > 0 && len(matches) > limit {
		matches = matches[:limit]
	}

	docs := Docs{}

	for _, match := range matches {
		doc, err := NewDoc(match, sourcePath)
		if err != nil {
			return nil, fmt.Errorf("could not open doc (%s): %w", match, err)
		}

		docs = append(docs, doc)
	}

	return docs, nil
}
