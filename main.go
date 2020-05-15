package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/paketo-buildpacks/packit"
)

func main() {
	switch filepath.Base(os.Args[0]) {
	case "detect":
		packit.Detect(Detect())
	case "build":
		packit.Build(Build())
	default:
		panic(fmt.Sprintf("unknown command: %s", filepath.Base(os.Args[0])))
	}
}
