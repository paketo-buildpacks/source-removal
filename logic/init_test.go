package logic_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitSourceRemoval(t *testing.T) {
	suite := spec.New("source-removal", spec.Report(report.Terminal{}))
	suite("Exclude", testExclude)
	suite("Include", testInclude)
	suite.Run(t)
}
