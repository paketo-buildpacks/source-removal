package sourceremoval_test

import (
	"testing"

	sourceremoval "github.com/ForestEckhardt/source-removal"
	"github.com/paketo-buildpacks/packit/v2"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testDetect(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		detect packit.DetectFunc
	)

	it.Before(func() {
		detect = sourceremoval.Detect()
	})

	it("passes detection", func() {
		result, err := detect(packit.DetectContext{})
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal(packit.DetectResult{}))
	})

}
