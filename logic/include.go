package logic

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func Include(workingDir, patterns string) error {
	// The following constructs a set of all the file paths that are required from a
	// globed file to exist and prepends the working directory onto all of
	// those permutation
	//
	// Example:
	// Input: "public/data/*"
	// Output: ["working-dir/public", "working-dir/public/data", "working-dir/public/data/*"]
	var globs = []string{workingDir}
	for _, glob := range filepath.SplitList(patterns) {
		dirs := strings.Split(glob, string(os.PathSeparator))
		for i := range dirs {
			globs = append(globs, filepath.Join(workingDir, filepath.Join(dirs[:i+1]...)))
		}
	}

	err := filepath.Walk(workingDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		var match bool
		for _, glob := range globs {
			match, err = filepath.Match(glob, path)
			if err != nil {
				return err
			}
			if match {
				// filepath.SkipDir is returned here because this is a glob that
				// specifies everything in a directroy should be included in the match
				// including subdirectories. If we get a match on such a glob we want to
				// ignore all other files in that directory because they are files we
				// either want to keep in an includes context.
				//
				// Example:
				// "public/data/*" matches "public/data/file" but does not match
				// "public/data/directory/file" we obviously want that directory to
				// remain in an includes context so we use filepath.SkipDir when we
				// detect "public/data/file" and the glob ends in "/*" which skips
				// scanning "public/data" directory because we know we want all of the
				// contents and don't want to go any deeper.
				if strings.HasSuffix(glob, fmt.Sprintf("%c*", os.PathSeparator)) {
					return filepath.SkipDir
				}
				break
			}
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
		return err
	}

	return nil
}
