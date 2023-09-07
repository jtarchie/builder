package builder_test

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jtarchie/builder"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/atomic"
)

var _ = Describe("Watcher", func() {
	When("a file gets updated", func() {
		It("executes the callback with the affected filename", func() {
			var foundFilename atomic.String

			sourceDir, err := os.MkdirTemp("", "")
			Expect(err).NotTo(HaveOccurred())

			watcher := builder.NewWatcher(sourceDir)

			go func() {
				defer GinkgoRecover()

				err := watcher.Execute(func(filename string) error {
					foundFilename.Store(filename)

					return nil
				})
				Expect(err).NotTo(HaveOccurred())
			}()

			Consistently(foundFilename.Load).Should(Equal(""))

			expectedFilename := filepath.Join(sourceDir, "file")

			err = os.WriteFile(expectedFilename, []byte(""), os.ModePerm)
			Expect(err).NotTo(HaveOccurred())

			Eventually(foundFilename.Load).Should(Equal(expectedFilename))
		})
	})

	When("the source path does not exists", func() {
		It("returns an error", func() {
			watcher := builder.NewWatcher("asdf")

			err := watcher.Execute(func(s string) error {
				return nil
			})
			Expect(err).To(HaveOccurred())
		})
	})

	When("the callback returns an error", func() {
		It("stops watching altogether", func() {
			var callbackCount atomic.Int32

			sourceDir, err := os.MkdirTemp("", "")
			Expect(err).NotTo(HaveOccurred())

			watcher := builder.NewWatcher(sourceDir)

			go func() {
				defer GinkgoRecover()

				err := watcher.Execute(func(filename string) error {
					callbackCount.Add(1)

					return fmt.Errorf("some error")
				})
				Expect(err).To(HaveOccurred())
			}()

			Consistently(callbackCount.Load).Should(BeEquivalentTo(0))

			for i := 0; i < 10; i++ {
				expectedFilename := filepath.Join(sourceDir, "file")

				err = os.WriteFile(expectedFilename, []byte(fmt.Sprintf("%d", i)), os.ModePerm)
				Expect(err).NotTo(HaveOccurred())
			}

			Consistently(callbackCount.Load).Should(BeEquivalentTo(1))
		})
	})
})
