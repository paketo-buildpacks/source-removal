package integration_test

import (
	"fmt"
	"testing"

	"github.com/cloudfoundry/dagger"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"

	. "github.com/onsi/gomega"
)

var (
	buildpack          string
	buildPlanBuildpack string
)

func TestIntegration(t *testing.T) {
	var Expect = NewWithT(t).Expect

	bpDir, err := dagger.FindBPRoot()
	Expect(err).NotTo(HaveOccurred())

	buildpack, err = dagger.PackageBuildpack(bpDir)
	Expect(err).ToNot(HaveOccurred())
	buildpack = fmt.Sprintf("%s.tgz", buildpack)

	buildPlanBuildpack, err = dagger.GetLatestCommunityBuildpack("ForestEckhardt", "build-plan")
	Expect(err).NotTo(HaveOccurred())

	defer func() {
		dagger.DeleteBuildpack(buildpack)
		dagger.DeleteBuildpack(buildPlanBuildpack)
	}()

	suite := spec.New("Integration", spec.Parallel(), spec.Report(report.Terminal{}))
	suite("Default", testDefault)
	suite.Run(t)
}
