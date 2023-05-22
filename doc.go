package builder

import (
	"bytes"
	"fmt"
	"os"
	"regexp"

	"github.com/adrg/frontmatter"
)

type Doc struct {
	contents string
	filename string
	metadata *DocMetadata
}

type DocMetadata struct {
	Title       string `yaml:"title"`
	Description string `yaml:"description"`
}

var titleFromHeader = regexp.MustCompile(`(?m)^#\s+(.*)$`)

func NewDoc(
	filename string,
) (*Doc, error) {
	metadata, contents, err := parseDoc(filename)
	if err != nil {
		return nil, fmt.Errorf("could not parse doc (%s): %w", filename, err)
	}

	return &Doc{
		contents: contents,
		filename: filename,
		metadata: metadata,
	}, nil
}

func (d *Doc) Metadata() *DocMetadata {
	return d.metadata
}

func (d *Doc) Title() string {
	if d.metadata.Title != "" {
		return d.metadata.Title
	}

	matches := titleFromHeader.FindAllStringSubmatch(d.contents, 1)
	if len(matches) == 0 {
		return ""
	}

	d.metadata.Title = matches[0][1]

	return d.metadata.Title
}

func (d *Doc) Description() string {
	return d.metadata.Description
}

func (d *Doc) Contents() string {
	return d.contents
}

func (d *Doc) Filename() string {
	return d.filename
}

func parseDoc(filename string) (*DocMetadata, string, error) {
	contents, err := os.ReadFile(filename)
	if err != nil {
		return nil, "", fmt.Errorf("could not read file (%s): %w", filename, err)
	}

	metadata := &DocMetadata{}

	leftovers, err := frontmatter.Parse(bytes.NewReader(contents), &metadata)
	if err != nil {
		return nil, "", fmt.Errorf("could not find front matter (%s): %w", filename, err)
	}

	return metadata, string(leftovers), nil
}
