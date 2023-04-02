package builder_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/jtarchie/builder"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
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

		cli = &builder.CLI{
			SourcePath:     sourcePath,
			BuildPath:      buildPath,
			LayoutFilename: "layout.html",
		}
	})

	When("a layout and assets directory exists", func() {
		BeforeEach(func() {
			createLayout()
			createFile("index.md", "---\ntitle: required\n---\nsome text")
			createFile("public/404.html", "404 page")
		})

		It("renders all files successfully", func() {
			err := cli.Run()
			Expect(err).NotTo(HaveOccurred())

			contents := readFile("index.html")
			Expect(contents).To(ContainSubstring("some text"))
			Expect(contents).To(ContainSubstring("<title>required</title>"))

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
			err := cli.Run()
			Expect(err).NotTo(HaveOccurred())

			contents := readFile("index.html")
			Expect(contents).To(ContainSubstring("<title>Some Title</title>"))
			Expect(contents).To(ContainSubstring("<description>Some Description</description>"))
		})
	})

	When("rendering documents without frontmatter", func() {
		It("errors on no title", func() {
			createLayout()
			createFile("index.md", "some text")

			err := cli.Run()
			Expect(err).To(HaveOccurred())
		})

		It("uses H1 for the title", func() {
			createLayout()
			createFile("index.md", "# some title")

			err := cli.Run()
			Expect(err).NotTo(HaveOccurred())

			contents := readFile("index.html")
			Expect(contents).To(ContainSubstring("<title>some title</title>"))
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

			createFile("posts/index.md", `# IGNORE ME`)

			createFile("index.md", `
---
title: required
---
{{range $doc := iterDocs "posts/" 3}}
* [{{$doc.Title}}]({{$doc.Path}}) {{$doc.Basename}}
{{end}}
			`)
			createFile("index-all.md", `
---
title: required
---
{{range $doc := iterDocs "posts/" 0}}
* [{{$doc.Title}}]({{$doc.Path}})
{{end}}
			`)
		})

		It("looks for the latest posts within a directory", func() {
			err := cli.Run()
			Expect(err).NotTo(HaveOccurred())

			contents := readFile("index.html")
			Expect(contents).ToNot(ContainSubstring(`IGNORE ME`))
			Expect(contents).To(ContainSubstring(`<a href="/posts/10.html">some 10 title</a> 10`))
			Expect(contents).To(ContainSubstring(`<a href="/posts/09.html">some 9 title</a>`))
			Expect(contents).To(ContainSubstring(`<a href="/posts/08.html">some 8 title</a>`))
			Expect(contents).NotTo(ContainSubstring(`<a href="/posts/07.html">some 7 title</a>`))

			contents = readFile("index-all.html")
			Expect(contents).To(ContainSubstring(`<a href="/posts/10.html">some 10 title</a>`))
			Expect(contents).To(ContainSubstring(`<a href="/posts/09.html">some 9 title</a>`))
			Expect(contents).To(ContainSubstring(`<a href="/posts/08.html">some 8 title</a>`))
			Expect(contents).To(ContainSubstring(`<a href="/posts/07.html">some 7 title</a>`))
			Expect(contents).To(ContainSubstring(`<a href="/posts/06.html">some 6 title</a>`))
			Expect(contents).To(ContainSubstring(`<a href="/posts/05.html">some 5 title</a>`))
			Expect(contents).To(ContainSubstring(`<a href="/posts/04.html">some 4 title</a>`))
			Expect(contents).To(ContainSubstring(`<a href="/posts/03.html">some 3 title</a>`))
			Expect(contents).To(ContainSubstring(`<a href="/posts/02.html">some 2 title</a>`))
			Expect(contents).To(ContainSubstring(`<a href="/posts/01.html">some 1 title</a>`))
		})
	})

	When("no layout exists", func() {
		It("fails with an error", func() {
			err := cli.Run()
			Expect(err).To(HaveOccurred())
		})
	})

	When("using Github emojis", func() {
		It("renders it", func() {
			createLayout()
			createFile("index.md", "---\ntitle: required\n---\n:running:")

			err := cli.Run()
			Expect(err).NotTo(HaveOccurred())

			contents := readFile("index.html")
			Expect(contents).NotTo(ContainSubstring(`running`))
			Expect(contents).To(ContainSubstring(`&#x1f3c3;`))
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

			err = cli.Run()
			Expect(err).NotTo(HaveOccurred())

			contents, err := os.ReadFile(filepath.Join(buildPath, "markdown.html"))
			Expect(err).NotTo(HaveOccurred())

			Expect(string(contents)).To(ContainSubstring(`<h2 id="h2-heading">h2 Heading</h2>`))
		})
	})
})
