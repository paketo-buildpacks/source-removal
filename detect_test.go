package main_test

import (
	"testing"

	main "github.com/ForestEckhardt/clear-source"
	"github.com/cloudfoundry/packit"
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

	it("returns a plan that provides and requires node", func() {
		result, err := detect(packit.DetectContext{
			WorkingDir: "/working-dir",
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(result.Plan).To(Equal(packit.BuildPlan{
			Provides: []packit.BuildPlanProvision{
				{Name: "no-source"},
			},
			Requires: []packit.BuildPlanRequirement{
				{Name: "no-source"},
			},
		}))
	})

}
