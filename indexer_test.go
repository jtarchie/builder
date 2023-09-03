package builder_test

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/jtarchie/builder"
)

var _ = Describe("Indexer", func() {
	It("indexes the HTML files and create an index.json", func() {
		// Create a temporary directory
		tmpDir, err := os.MkdirTemp("", "indexer")
		Expect(err).ToNot(HaveOccurred())

		// Create a sample HTML file in the temporary directory
		htmlContent := `
			<html>
				<head>
					<title>Test Title</title>
				</head>
				<body>
					Test Content
				</body>
			</html>
			`
		err = os.WriteFile(filepath.Join(tmpDir, "sample.html"), []byte(htmlContent), 0644)
		Expect(err).ToNot(HaveOccurred())

		indexer := builder.NewIndexer(tmpDir)

		err = indexer.Execute()
		Expect(err).ToNot(HaveOccurred())

		// Verify that index.json was created
		_, err = os.Stat(filepath.Join(tmpDir, "index.json"))
		Expect(err).ToNot(HaveOccurred())
	})
})
