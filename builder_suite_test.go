package builder_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/jtarchie/builder"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
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

		err := os.MkdirAll(filepath.Dir(fullPath), os.ModePerm)
		Expect(err).NotTo(HaveOccurred())

		err = os.WriteFile(fullPath, []byte(contents), os.ModePerm)
		Expect(err).NotTo(HaveOccurred())
	}

	readFile := func(filename string) *gbytes.Buffer {
		fullPath := filepath.Join(buildPath, filename)

		contents, err := os.ReadFile(fullPath)
		Expect(err).NotTo(HaveOccurred())

		return gbytes.BufferWithBytes(contents)
	}

	createLayout := func() {
		createFile("layout.html", `
			<html>
				<head>
					<title>{{.Doc.Title}}</title>
					<description>{{.Doc.Description}}</description>
				</head>
				<body>
				{{.RenderedPage}}
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
			createFile("index.md", "---\ntitle: Required Title\n---\nsome text")
			createFile("another.md", "---\ntitle: Some 😂 Title\n---\nsome text")
			createFile("public/404.html", "404 page")
		})

		It("renders all files successfully", func() {
			err := cli.Run()
			Expect(err).NotTo(HaveOccurred())

			contents := readFile("index.html")
			Eventually(contents).Should(gbytes.Say("<title>Required Title</title>"))
			Eventually(contents).Should(gbytes.Say("some text"))

			contents = readFile("another.html")
			Eventually(contents).Should(gbytes.Say("<title>Some 😂 Title</title>"))

			contents = readFile("another-some-title.html")
			Eventually(contents).Should(gbytes.Say("<title>Some 😂 Title</title>"))

			contents = readFile("404.html")
			Eventually(contents).Should(gbytes.Say("404 page"))
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
			Eventually(contents).Should(gbytes.Say("<title>Some Title</title>"))
			Eventually(contents).Should(gbytes.Say("<description>Some Description</description>"))
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
			Eventually(contents).Should(gbytes.Say("<title>some title</title>"))
		})
	})

	When("rendering a file with template functions", func() {
		BeforeEach(func() {
			createLayout()

			for i := 1; i <= 10; i++ {
				createFile(
					fmt.Sprintf("posts/%02d.md", i),
					fmt.Sprintf("---\ntitle: some %d 😂 title\n---\nsome text", i),
				)
			}

			createFile("posts/index.md", `# IGNORE ME`)

			createFile("index.md", `
---
title: required
---
{{range $doc := iterDocs "posts/" 3}}
* [{{$doc.Title}}]({{$doc.Path}}) {{$doc.Basename}}
* [{{$doc.Title}}]({{$doc.SlugPath}}) {{$doc.Basename}}
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
			Eventually(contents).ShouldNot(gbytes.Say(`IGNORE ME`))
			Eventually(contents).Should(gbytes.Say(`<a href="/posts/10.html">some 10 😂 title</a> 10`))
			Eventually(contents).Should(gbytes.Say(`<a href="/posts/10-some-10-title.html">some 10 😂 title</a> 10`))
			Eventually(contents).Should(gbytes.Say(`<a href="/posts/09.html">some 9 😂 title</a>`))
			Eventually(contents).Should(gbytes.Say(`<a href="/posts/08.html">some 8 😂 title</a>`))
			Eventually(contents).ShouldNot(gbytes.Say(`<a href="/posts/07.html">some 7 😂 title</a>`))

			contents = readFile("index-all.html")
			Eventually(contents).Should(gbytes.Say(`<a href="/posts/10.html">some 10 😂 title</a>`))
			Eventually(contents).Should(gbytes.Say(`<a href="/posts/09.html">some 9 😂 title</a>`))
			Eventually(contents).Should(gbytes.Say(`<a href="/posts/08.html">some 8 😂 title</a>`))
			Eventually(contents).Should(gbytes.Say(`<a href="/posts/07.html">some 7 😂 title</a>`))
			Eventually(contents).Should(gbytes.Say(`<a href="/posts/06.html">some 6 😂 title</a>`))
			Eventually(contents).Should(gbytes.Say(`<a href="/posts/05.html">some 5 😂 title</a>`))
			Eventually(contents).Should(gbytes.Say(`<a href="/posts/04.html">some 4 😂 title</a>`))
			Eventually(contents).Should(gbytes.Say(`<a href="/posts/03.html">some 3 😂 title</a>`))
			Eventually(contents).Should(gbytes.Say(`<a href="/posts/02.html">some 2 😂 title</a>`))
			Eventually(contents).Should(gbytes.Say(`<a href="/posts/01.html">some 1 😂 title</a>`))
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
			Eventually(contents).Should(gbytes.Say(`&#x1f3c3;`))
			Eventually(contents).ShouldNot(gbytes.Say(`running`))
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

			buffer := gbytes.BufferWithBytes(contents)

			Eventually(buffer).Should(gbytes.Say(`<h2 id="h2-heading">h2 Heading</h2>`))
		})
	})
})
