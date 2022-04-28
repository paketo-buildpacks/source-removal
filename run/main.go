package main

import (
	"github.com/paketo-buildpacks/packit/v2"
	sourceremoval "github.com/paketo-buildpacks/source-removal"
)

func main() {
	packit.Run(
		sourceremoval.Detect(),
		sourceremoval.Build())
}
