package sourceremoval

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/paketo-buildpacks/packit"
)

func Build() packit.BuildFunc {
	return func(context packit.BuildContext) (packit.BuildResult, error) {
		var envGlobs []string
		if val, ok := os.LookupEnv("BP_INCLUDE_FILES"); ok {
			envGlobs = append(envGlobs, filepath.SplitList(val)...)
		}

		// The following constructs a set of all the file paths that are required from a
		// globed file to exist and prepends the working directory onto all of
		// those permutation
		//
		// Example:
		// Input: "public/data/*"
		// Output: ["working-dir/public", "working-dir/public/data", "working-dir/public/data/*"]
		var globs = []string{context.WorkingDir}
		for _, glob := range envGlobs {
			dirs := strings.Split(glob, string(os.PathSeparator))
			for i := range dirs {
				globs = append(globs, filepath.Join(context.WorkingDir, filepath.Join(dirs[:i+1]...)))
			}
		}

		err := filepath.Walk(context.WorkingDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			match, err := matchingGlob(path, globs)
			if err != nil {
				return err
			}

			if !match {
				err := os.RemoveAll(path)
				if err != nil {
					return err
				}

				if info.IsDir() {
					return filepath.SkipDir
				}
			}

			return nil
		})

		if err != nil {
			return packit.BuildResult{}, err
		}

		return packit.BuildResult{}, nil
	}
}

func matchingGlob(path string, globs []string) (bool, error) {
	for _, glob := range globs {
		match, err := filepath.Match(glob, path)
		if err != nil {
			return false, err
		}

		if match {
			// filepath.SkipDir is returned here because this is a glob that
			// specifies everything in a directroy. If we get a match on such
			// a glob we want to ignore all other files in that directory because
			// they are files we want to keep and the glob will not work
			// if it enters that directory
			if strings.HasSuffix(glob, fmt.Sprintf("%c*", os.PathSeparator)) {
				return true, filepath.SkipDir
			}
			return true, nil
		}
	}

	return false, nil
}
