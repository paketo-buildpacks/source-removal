package main

import (
	sourceremoval "github.com/ForestEckhardt/source-removal"
	"github.com/paketo-buildpacks/packit/v2"
)

func main() {
	packit.Run(
		sourceremoval.Detect(),
		sourceremoval.Build())
}
