package rag_test

import (
	"testing"

	"github.com/jtarchie/builder/rag"
	"github.com/onsi/gomega"
)

func TestError(t *testing.T) {
	assert := gomega.NewWithT(t)

	_, err := rag.New("/asdfasdf/asdfasdfasd", nil)
	assert.Expect(err).To(gomega.HaveOccurred())
}
