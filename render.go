package builder

import (
	"bytes"
	// embed file assets.
	_ "embed"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/Masterminds/sprig/v3"
	"github.com/bmatcuk/doublestar/v4"
	"github.com/gorilla/feeds"
	"github.com/gosimple/slug"
	"github.com/microcosm-cc/bluemonday"
	cp "github.com/otiai10/copy"
	"github.com/sabloger/sitemap-generator/smg"
	"github.com/tdewolff/minify"
	mHTML "github.com/tdewolff/minify/html"
	"github.com/yuin/goldmark"
	emoji "github.com/yuin/goldmark-emoji"
	highlighting "github.com/yuin/goldmark-highlighting"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"go.abhg.dev/goldmark/anchor"
	"go.abhg.dev/goldmark/mermaid"
	"golang.org/x/sync/errgroup"
)

type Render struct {
	assetsPath string
	baseURL    string
	buildPath  string
	converter  goldmark.Markdown
	layoutPath string
	sourcePath string
}

var errMissingTitle = errors.New("missing title in metadata or h1")

func NewRender(
	layoutPath string,
	sourcePath string,
	assetsPath string,
	buildPath string,
	baseURL string,
) *Render {
	return &Render{
		assetsPath: assetsPath,
		baseURL:    baseURL,
		buildPath:  buildPath,
		layoutPath: layoutPath,
		sourcePath: sourcePath,
		converter: goldmark.New(
			goldmark.WithRendererOptions(
				html.WithXHTML(),
				html.WithUnsafe(),
			),
			goldmark.WithExtensions(
				extension.GFM,
				emoji.Emoji,
				&mermaid.Extender{},
				highlighting.Highlighting,
				extension.DefinitionList,
				extension.Footnote,
				extension.Typographer,
				&anchor.Extender{
					Texter:   anchor.Text("#"),
					Position: anchor.Before,
				},
			),
			goldmark.WithParserOptions(
				parser.WithAutoHeadingID(),
				parser.WithAttribute(),
			),
		),
	}
}

//nolint:funlen,cyclop
func (r *Render) Execute(
	docsGlob string,
	feedGlob string,
) error {
	err := r.copyAssets()
	if err != nil {
		return fmt.Errorf("copying assets issue: %w", err)
	}

	docs, err := NewDocs(r.sourcePath, docsGlob, 0, false)
	if err != nil {
		return fmt.Errorf("could not glob markdown files: %w", err)
	}

	contents, err := readFile(r.layoutPath)
	if err != nil {
		return fmt.Errorf("could not read layout: %w", err)
	}

	layout, err := template.New(r.layoutPath).Parse(contents)
	if err != nil {
		return fmt.Errorf("could not parse layout template (%s): %w", r.layoutPath, err)
	}

	funcMap := template.FuncMap{
		"iterDocs": func(path string, limit int) (Docs, error) {
			pattern := filepath.Join(r.sourcePath, path, "*.md")
			docs, err := NewDocs(r.sourcePath, pattern, limit, true)
			if err != nil {
				return nil, fmt.Errorf("could not load docs: %w", err)
			}

			return docs, nil
		},
	}

	maxRenders := 10
	group := &errgroup.Group{}

	group.SetLimit(maxRenders)

	for _, doc := range docs {
		doc := doc

		group.Go(func() error {
			err := r.renderDocument(doc, funcMap, layout)
			if err != nil {
				return fmt.Errorf("rendering template issue: %w", err)
			}

			return nil
		})
	}

	if r.baseURL != "" {
		group.Go(func() error {
			err = r.generateFeeds(docs, feedGlob, funcMap)
			if err != nil {
				return fmt.Errorf("could not render feeds: %w", err)
			}

			return nil
		})
	}

	err = group.Wait()
	if err != nil {
		return fmt.Errorf("could not render: %w", err)
	}

	return nil
}

//nolint:funlen
func (r *Render) generateFeeds(
	docs Docs,
	feedGlob string,
	funcMap template.FuncMap,
) error {
	now := time.Now().UTC()

	feed := &feeds.Feed{
		Title:       r.baseURL,
		Link:        &feeds.Link{Href: r.baseURL},
		Description: fmt.Sprintf("feed for %s", r.baseURL),
		Created:     now,
	}

	sitemap := smg.NewSitemap(true)
	sitemap.SetOutputPath(r.buildPath)
	sitemap.SetLastMod(&now)
	sitemap.SetCompress(false)
	sitemap.SetHostname(r.baseURL)

	sanitizer := bluemonday.UGCPolicy()

	for _, doc := range docs {
		if matched, _ := doublestar.Match(feedGlob, doc.Filename()); !matched {
			continue
		}

		modifiedTime := doc.Timespec.ModTime().UTC()
		createdTime := modifiedTime

		if doc.Timespec.HasBirthTime() {
			createdTime = doc.Timespec.BirthTime().UTC()
		}

		docURL, _ := url.JoinPath(r.baseURL, doc.Path())

		err := sitemap.Add(&smg.SitemapLoc{
			Loc:        docURL,
			LastMod:    &modifiedTime,
			ChangeFreq: smg.Always,
		})
		if err != nil {
			return fmt.Errorf("could not add file %q to sitemap: %w", doc.Filename(), err)
		}

		contents, _ := r.renderMarkdownFromDoc(doc, funcMap)

		feed.Items = append(feed.Items, &feeds.Item{
			Id:    docURL,
			Title: doc.Title(),
			Link: &feeds.Link{
				Href: docURL,
			},
			Description: doc.Description(),
			Updated:     modifiedTime,
			Created:     createdTime,
			Content:     sanitizer.Sanitize(contents),
		})
	}

	_, err := sitemap.Save()
	if err != nil {
		return fmt.Errorf("could not save sitemap: %w", err)
	}

	atomFeed, err := feed.ToAtom()
	if err != nil {
		return fmt.Errorf("could not generate atom feed: %w", err)
	}

	err = os.WriteFile(filepath.Join(r.buildPath, "atom.xml"), []byte(atomFeed), os.ModePerm)
	if err != nil {
		return fmt.Errorf("could not write atom feed: %w", err)
	}

	rssFeed, err := feed.ToRss()
	if err != nil {
		return fmt.Errorf("could not generate rss feed: %w", err)
	}

	err = os.WriteFile(filepath.Join(r.buildPath, "rss.xml"), []byte(rssFeed), os.ModePerm)
	if err != nil {
		return fmt.Errorf("could not write rss feed: %w", err)
	}

	return nil
}

func (r *Render) copyAssets() error {
	err := os.RemoveAll(r.buildPath)
	if err != nil {
		return fmt.Errorf("could not remove build path (%s): %w", r.buildPath, err)
	}

	err = os.MkdirAll(r.buildPath, os.ModePerm)
	if err != nil {
		return fmt.Errorf("could not create build path (%s): %w", r.buildPath, err)
	}

	assetsPath := r.assetsPath

	if _, err := os.Stat(assetsPath); !os.IsNotExist(err) {
		opt := cp.Options{
			Skip: func(info os.FileInfo, src, dest string) (bool, error) {
				// skip dot files
				return strings.HasPrefix(filepath.Base(src), "."), nil
			},
		}

		err = cp.Copy(assetsPath, r.buildPath, opt)
		if err != nil {
			return fmt.Errorf("could not copy assets contents (%s): %w", assetsPath, err)
		}
	}

	return nil
}

func (r *Render) renderMarkdownFromDoc(doc *Doc, funcMap template.FuncMap) (string, error) {
	filename := doc.Filename()

	if doc.Title() == "" {
		return "", fmt.Errorf("could not determine title (%s): %w", filename, errMissingTitle)
	}

	markdown, err := template.
		New(filename).
		Funcs(funcMap).
		Funcs(sprig.FuncMap()).
		Parse(doc.Contents())
	if err != nil {
		return "", fmt.Errorf("could not parse markdown template (%s): %w", r.layoutPath, err)
	}

	markdownWriter, renderedWriter := &bytes.Buffer{}, &bytes.Buffer{}

	err = markdown.Execute(markdownWriter, nil)
	if err != nil {
		return "", fmt.Errorf("could not render markdown template (%s): %w", filename, err)
	}

	err = r.converter.Convert(markdownWriter.Bytes(), renderedWriter)
	if err != nil {
		return "", fmt.Errorf("could not convert file (%s): %w", filename, err)
	}

	return renderedWriter.String(), nil
}

func (r *Render) renderDocument(doc *Doc, funcMap template.FuncMap, layout *template.Template) error {
	renderedMarkdown, err := r.renderMarkdownFromDoc(doc, funcMap)
	if err != nil {
		return fmt.Errorf("could not render markdown for doc: %w", err)
	}

	layoutWriter := &bytes.Buffer{}

	err = layout.Execute(layoutWriter, map[string]any{
		"Doc": doc,

		"RenderedPage": renderedMarkdown,
	})
	if err != nil {
		return fmt.Errorf("could not render layout template (%s): %w", doc.Filename(), err)
	}

	withoutSlugFilename := strings.Replace(doc.Filename(), r.sourcePath, r.buildPath, 1)
	withoutSlugFilename = strings.Replace(withoutSlugFilename, ".md", ".html", 1)
	filenames := []string{withoutSlugFilename}

	if !strings.Contains(withoutSlugFilename, "index.html") {
		withSlugFilename := strings.Replace(withoutSlugFilename, ".html", "-"+slug.Make(doc.Title())+".html", 1)
		filenames = append(filenames, withSlugFilename)
	}

	err = writeHTMLFiles(filenames, layoutWriter.String())
	if err != nil {
		return fmt.Errorf("could write new template: %w", err)
	}

	return nil
}

func readFile(filename string) (string, error) {
	contents, err := os.ReadFile(filename)
	if err != nil {
		return "", fmt.Errorf("could not read file (%s): %w", filename, err)
	}

	return string(contents), nil
}

func writeHTMLFiles(filenames []string, contents string) error {
	dirPath := filepath.Dir(filenames[0])

	err := os.MkdirAll(dirPath, os.ModePerm)
	if err != nil {
		return fmt.Errorf("could not create path (%s): %w", dirPath, err)
	}

	writer := &strings.Builder{}
	minifier := &mHTML.Minifier{
		KeepDocumentTags: true,
	}

	err = minifier.Minify(&minify.M{}, writer, strings.NewReader(contents), nil)
	if err != nil {
		return fmt.Errorf("could not minify: %w", err)
	}

	for _, filename := range filenames {
		err = os.WriteFile(filename, []byte(writer.String()), os.ModePerm)
		if err != nil {
			return fmt.Errorf("could not write path (%s): %w", filename, err)
		}
	}

	return nil
}
