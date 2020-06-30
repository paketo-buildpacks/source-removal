package main

import (
	sourceremoval "github.com/ForestEckhardt/source-removal"
	"github.com/paketo-buildpacks/packit"
)

func main() {
	packit.Run(
		sourceremoval.Detect(),
		sourceremoval.Build())
}
