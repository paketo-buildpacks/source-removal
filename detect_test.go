package main_test

import (
	"testing"

	main "github.com/ForestEckhardt/source-removal"
	"github.com/paketo-buildpacks/packit"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testDetect(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		detect packit.DetectFunc
	)

	it.Before(func() {
		detect = main.Detect()
	})

	it("passes detection", func() {
		result, err := detect(packit.DetectContext{})
		Expect(err).NotTo(HaveOccurred())
		Expect(result.Plan).To(Equal(packit.BuildPlan{
			Provides: []packit.BuildPlanProvision{
				{Name: "source-removal"},
			},
		}))
	})

}
