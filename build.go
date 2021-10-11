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

		switch {
		case includeSet:
			err := includeFiles(includeVal, context.WorkingDir)
			if err != nil {
				return packit.BuildResult{}, err
			}
		case excludeSet:
			err := excludeFiles(excludeVal, context.WorkingDir)
			if err != nil {
				return packit.BuildResult{}, err
			}
		default:
			err := removeAllFiles(context.WorkingDir)
			if err != nil {
				return packit.BuildResult{}, err
			}
		}

		return packit.BuildResult{}, nil
	}
}

func includeFiles(includeVal, workingDir string) error {
	var envGlobs []string
	envGlobs = append(envGlobs, filepath.SplitList(includeVal)...)

	// The following constructs a set of all the file paths that are required from a
	// globed file to exist and prepends the working directory onto all of
	// those permutation
	//
	// Example:
	// Input: "public/data/*"
	// Output: ["working-dir/public", "working-dir/public/data", "working-dir/public/data/*"]
	var globs = []string{workingDir}
	for _, glob := range envGlobs {
		dirs := strings.Split(glob, string(os.PathSeparator))
		for i := range dirs {
			globs = append(globs, filepath.Join(workingDir, filepath.Join(dirs[:i+1]...)))
		}
	}

	return filepath.Walk(workingDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		match, err := matchingGlob(path, globs, true)
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
}

func excludeFiles(excludeVal, workingDir string) error {
	var globs []string
	for _, g := range filepath.SplitList(excludeVal) {
		globs = append(globs, filepath.Join(workingDir, g))
	}

	return filepath.Walk(workingDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		match, err := matchingGlob(path, globs, false)
		if err != nil {
			return err
		}

		if match {
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

func matchingGlob(path string, globs []string, includeSet bool) (bool, error) {
	for _, glob := range globs {
		match, err := filepath.Match(glob, path)
		if err != nil {
			return false, err
		}

		if match {
			if includeSet {
				// filepath.SkipDir is returned here because this is a glob that
				// specifies everything in a directroy. If we get a match on such
				// a glob we want to ignore all other files in that directory because
				// they are files we want to keep and the glob will not work
				// if it enters that directory
				if strings.HasSuffix(glob, fmt.Sprintf("%c*", os.PathSeparator)) {
					return true, filepath.SkipDir
				}
			}
			return true, nil
		}
	}

	return false, nil
}
