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

		switch {
		case includeSet:
			err := includeFiles(includeVal, context.WorkingDir)
			if err != nil {
				return packit.BuildResult{}, err
			}
			fallthrough
		case excludeSet:
			err := excludeFiles(excludeVal, context.WorkingDir)
			if err != nil {
				return packit.BuildResult{}, err
			}
		default:
			err := removeAllFiles(context.WorkingDir)
			return packit.BuildResult{}, err
		}

		return packit.BuildResult{}, nil
	}
}

func includeFiles(includeVal, workingDir string) error {
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

	err := filepath.Walk(
		workingDir,
		generateWalkFunc(
			globs,
			func(match bool) bool {
				// If the match is true we do not want to remove the file or directoy in
				// the include case
				return !match
			},
		))

	return err
}

func excludeFiles(excludeVal, workingDir string) error {
	var globs []string
	for _, g := range filepath.SplitList(excludeVal) {
		globs = append(globs, filepath.Join(workingDir, g))
	}

	err := filepath.Walk(
		workingDir,
		generateWalkFunc(
			globs,
			func(match bool) bool {
				// If the match is true we do want to remove the file or directoy in
				// the exclude case
				return match
			},
		))

	return err
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

func generateWalkFunc(globs []string, removalCheckFunc func(bool) bool) func(path string, info os.FileInfo, err error) error {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		match, err := matchingGlob(path, globs)
		if errors.Is(err, filepath.SkipDir) {
			// If we don't want the file removed but the matcher has returned a
			// filepath.SkipDir then then we should skip this directory completely
			// because the glob ended in '/*' which means that everything in the
			// current directory is something that should be kept.
			if !removalCheckFunc(match) {
				return err
			}
		} else if err != nil {
			return err
		}

		if removalCheckFunc(match) {
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
}

func matchingGlob(path string, globs []string) (bool, error) {
	for _, glob := range globs {
		match, err := filepath.Match(glob, path)
		if err != nil {
			return false, err
		}

		if match {
			// filepath.SkipDir is returned here because this is a glob that
			// specifies everything in a directroy should be included in the match
			// including subdirectories. If we get a match on such a glob we want to
			// ignore all other files in that directory because they are files we
			// either want to keep in an includes context or they will be deleted on
			// match in the exclude context.
			//
			// Example:
			// "public/data/*" matches "public/data/file" but does not match
			// "public/data/directory/file" we obviously want that directory to
			// remain in an includes context so we use filepath.SkipDir when we
			// detect "public/data/file" and the glob ends in "/*" which skips
			// scanning "public/data" directory because we know we want all of the
			// contents and don't want to go any deeper.
			if strings.HasSuffix(glob, fmt.Sprintf("%c*", os.PathSeparator)) {
				return true, filepath.SkipDir
			}
			return true, nil
		}
	}

	return false, nil
}
