package sourceremoval_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitSourceRemoval(t *testing.T) {
	suite := spec.New("source-removal", spec.Report(report.Terminal{}))
	suite("Build", testBuild)
	suite("Detect", testDetect)
	suite("Exclude", testExclude)
	suite("Include", testInclude)
	suite.Run(t)
}
