package sourceremoval

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/paketo-buildpacks/packit"
)

func Build() packit.BuildFunc {
	return func(context packit.BuildContext) (packit.BuildResult, error) {
		var globs []string

		for _, entry := range context.Plan.Entries {
			existingGlobs, ok := entry.Metadata["keep"]
			if !ok {
				continue
			}

			interfaceGlobs, ok := existingGlobs.([]interface{})
			if !ok {
				return packit.BuildResult{}, errors.New("Error: keep field needs to be a list of strings")
			}

			var rawGlobs []string
			for _, interfaceGlob := range interfaceGlobs {
				rawGlob, _ := interfaceGlob.(string)
				rawGlobs = append(rawGlobs, rawGlob)
			}

			// The following constructs a set of all the file paths that are required from a
			// globed file to exist and prepends the working directory onto all of
			// those permutation
			//
			// Example:
			// Input: "public/data/*"
			// Output: ["working-dir/public", "working-dir/public/data", "working-dir/public/data/*"]
			for _, glob := range rawGlobs {
				dirs := strings.Split(glob, string(os.PathSeparator))
				for i := 1; i <= len(dirs); i++ {
					globs = append(globs, filepath.Join(context.WorkingDir, filepath.Join(dirs[:i]...)))
				}
			}
		}

		removalFunc := func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Prevents up from deleting the working directory ... Thanks filepath.Walk
			if path == context.WorkingDir {
				return nil
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
		}

		err := filepath.Walk(context.WorkingDir, removalFunc)
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
			// filepath.SkipDir is returned here becase this is a glob that
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
