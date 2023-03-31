package builder

import (
	"fmt"
	"sort"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/samber/lo"
)

type Docs []*Doc

func NewDocs(
	pattern string,
	limit int,
) (Docs, error) {
	matches, err := doublestar.FilepathGlob(pattern)
	if err != nil {
		return nil, fmt.Errorf("could not matches (%q): %w", pattern, err)
	}

	matches = lo.Filter(matches, func(match string, _ int) bool {
		return !strings.HasSuffix(match, "index.md")
	})

	sort.Strings(matches)
	sort.Sort(sort.Reverse(sort.StringSlice(matches)))

	if limit > 0 && len(matches) > limit {
		matches = matches[:limit]
	}

	docs := Docs{}

	for _, match := range matches {
		doc, err := NewDoc(match)
		if err != nil {
			return nil, fmt.Errorf("could not open doc (%s): %w", match, err)
		}

		docs = append(docs, doc)
	}

	return docs, nil
}
