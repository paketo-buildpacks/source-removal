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
		includeVal, includeSet := os.LookupEnv("BP_INCLUDE_FILES")
		excludeVal, excludeSet := os.LookupEnv("BP_EXCLUDE_FILES")

		if includeSet && excludeSet {
			return packit.BuildResult{}, errors.New("BP_INCLUDE_FILES and BP_EXCLUDE_FILES cannot be set at the same time")
		}

		var globs []string
		switch {
		case includeSet:
			globs = includeFiles(includeVal, context.WorkingDir)
		case excludeSet:
			globs = excludeFiles(excludeVal, context.WorkingDir)
		default:
			err := removeAllFiles(context.WorkingDir)
			return packit.BuildResult{}, err
		}

		err := filepath.Walk(context.WorkingDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			match, glob, err := matchingGlob(path, globs)
			if err != nil {
				return err
			}

			if match && includeSet {
				// filepath.SkipDir is returned here because this is a glob that
				// specifies everything in a directroy. If we get a match on such
				// a glob we want to ignore all other files in that directory because
				// they are files we want to keep and the glob will not work
				// if it enters that directory any further.
				//
				// Example:
				// "public/data/*" matches "public/data/file" but does not
				// match "public/data/directory/file" we obviously want that directory
				// to remain so we use filepath.SkipDir when we detect
				// "public/data/file" and the glob ends in "/*" which skips scanning
				// "public/data" directory because we know we want all of the contents
				// and don't want to go any deeper.
				if strings.HasSuffix(glob, fmt.Sprintf("%c*", os.PathSeparator)) {
					return filepath.SkipDir
				}
			}

			if (match && excludeSet) || (!match && includeSet) {
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

func includeFiles(includeVal, workingDir string) []string {
	// The following constructs a set of all the file paths that are required from a
	// globed file to exist and prepends the working directory onto all of
	// those permutation
	//
	// Example:
	// Input: "public/data/*"
	// Output: ["working-dir/public", "working-dir/public/data", "working-dir/public/data/*"]
	var globs = []string{workingDir}
	for _, glob := range filepath.SplitList(includeVal) {
		dirs := strings.Split(glob, string(os.PathSeparator))
		for i := range dirs {
			globs = append(globs, filepath.Join(workingDir, filepath.Join(dirs[:i+1]...)))
		}
	}

	return globs

}

func excludeFiles(excludeVal, workingDir string) []string {
	var globs []string
	for _, g := range filepath.SplitList(excludeVal) {
		globs = append(globs, filepath.Join(workingDir, g))
	}

	return globs
}

func removeAllFiles(workingDir string) error {
	files, err := os.ReadDir(workingDir)
	if err != nil {
		return err
	}

	for _, f := range files {
		err = os.RemoveAll(filepath.Join(workingDir, f.Name()))
		if err != nil {
			return err
		}
	}

	return nil
}

func matchingGlob(path string, globs []string) (bool, string, error) {
	for _, glob := range globs {
		match, err := filepath.Match(glob, path)
		if err != nil {
			return false, glob, err
		}

		if match {
			return true, glob, nil
		}
	}

	return false, "", nil
}
