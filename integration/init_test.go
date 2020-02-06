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
	bpDir, nosourceURI, goURI, goModURI string
)

func TestIntegration(t *testing.T) {

	var (
		Expect = NewWithT(t).Expect
		err    error
	)

	bpDir, err = dagger.FindBPRoot()
	Expect(err).NotTo(HaveOccurred())

	nosourceURI, err = dagger.PackageBuildpack(bpDir)
	Expect(err).ToNot(HaveOccurred())
	nosourceURI = fmt.Sprintf("%s.tgz", nosourceURI)

	defer dagger.DeleteBuildpack(nosourceURI)

	suite := spec.New("Integration", spec.Parallel(), spec.Report(report.Terminal{}))
	suite("SimpleApp", testSimpleApp)
	dagger.SyncParallelOutput(func() { suite.Run(t) })
}
