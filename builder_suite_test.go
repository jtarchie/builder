package builder_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/jtarchie/builder"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/zap"
)

func TestBuilder(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Builder Suite")
}

var _ = Describe("Builder", func() {
	var (
		sourcePath string
		buildPath  string
		cli        *builder.CLI
		logger     *zap.Logger
	)

	createFile := func(filename, contents string) {
		fullPath := filepath.Join(sourcePath, filename)

		err := os.MkdirAll(filepath.Dir(fullPath), 0777)
		Expect(err).NotTo(HaveOccurred())

		err = os.WriteFile(fullPath, []byte(contents), 0777)
		Expect(err).NotTo(HaveOccurred())
	}

	readFile := func(filename string) string {
		fullPath := filepath.Join(buildPath, filename)

		contents, err := os.ReadFile(fullPath)
		Expect(err).NotTo(HaveOccurred())

		return string(contents)
	}

	createLayout := func() {
		createFile("layout.html", `
			<html>
				<head>
					<title>{{.Title}}</title>
					<description>{{.Description}}</description>
				</head>
				<body>
				{{.Page}}
				</body>
			</html>
		`)
	}

	BeforeEach(func() {
		var err error

		sourcePath, err = os.MkdirTemp("", "")
		Expect(err).NotTo(HaveOccurred())

		buildPath, err = os.MkdirTemp("", "")
		Expect(err).NotTo(HaveOccurred())

		logger, err = zap.NewDevelopment()
		Expect(err).NotTo(HaveOccurred())

		cli = &builder.CLI{
			SourcePath:     sourcePath,
			BuildPath:      buildPath,
			LayoutFilename: "layout.html",
		}
	})

	When("a layout and assets directory exists", func() {
		BeforeEach(func() {
			createLayout()
			createFile("index.md", "---\n---\nsome text")
			createFile("public/404.html", "404 page")
		})

		It("renders all files successfully", func() {
			err := cli.Run(logger)
			Expect(err).NotTo(HaveOccurred())

			contents := readFile("index.html")
			Expect(contents).To(ContainSubstring("some text"))
			Expect(contents).To(ContainSubstring("<title></title>"))

			contents = readFile("404.html")
			Expect(contents).To(ContainSubstring("404 page"))
		})
	})

	When("rendering documents with frontmatter", func() {
		BeforeEach(func() {
			createLayout()
			createFile("index.md", "---\ntitle: Some Title\ndescription: Some Description\n---\nsome text")
		})

		It("renders a title and description", func() {
			err := cli.Run(logger)
			Expect(err).NotTo(HaveOccurred())

			contents := readFile("index.html")
			Expect(contents).To(ContainSubstring("<title>Some Title</title>"))
			Expect(contents).To(ContainSubstring("<description>Some Description</description>"))
		})
	})

	When("rendering a file with template functions", func() {
		BeforeEach(func() {
			createLayout()

			for i := 1; i <= 10; i++ {
				createFile(
					fmt.Sprintf("posts/%02d.md", i),
					fmt.Sprintf("---\ntitle: some %d title\n---\nsome text", i),
				)
			}

			createFile("index.md", `
---
---
{{range $doc := iterDocs "posts/" 3}}
* [{{$doc.Title}}]({{$doc.Path}})
{{end}}
			`)
		})

		It("looks for the latest posts within a directory", func() {
			err := cli.Run(logger)
			Expect(err).NotTo(HaveOccurred())

			contents := readFile("index.html")
			Expect(contents).To(ContainSubstring(`<a href="/posts/10.html">some 10 title</a>`))
			Expect(contents).To(ContainSubstring(`<a href="/posts/09.html">some 9 title</a>`))
			Expect(contents).To(ContainSubstring(`<a href="/posts/08.html">some 8 title</a>`))
			Expect(contents).NotTo(ContainSubstring(`<a href="/posts/07.html">some 7 title</a>`))
		})
	})

	When("no layout exists", func() {
		It("fails with an error", func() {
			err := cli.Run(logger)
			Expect(err).To(HaveOccurred())
		})
	})

	When("using the example", func() {
		It("is successful", func() {
			examplePath, err := filepath.Abs(filepath.Join(".", "example"))
			Expect(err).NotTo(HaveOccurred())

			cli = &builder.CLI{
				SourcePath:     examplePath,
				BuildPath:      buildPath,
				LayoutFilename: "layout.html",
			}

			err = cli.Run(logger)
			Expect(err).NotTo(HaveOccurred())

			contents, err := os.ReadFile(filepath.Join(buildPath, "markdown.html"))
			Expect(err).NotTo(HaveOccurred())

			Expect(string(contents)).To(ContainSubstring(`<h2 id="h2-heading">h2 Heading</h2>`))
		})
	})
})
