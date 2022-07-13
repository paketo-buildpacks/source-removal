package logic

import (
	"os"
	"path/filepath"
)

func Exclude(workingDir, patterns string) error {
	var globs []string
	for _, glob := range filepath.SplitList(patterns) {
		globs = append(globs, filepath.Join(workingDir, glob))
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
				break
			}
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
	if err != nil {
		return err
	}

	return nil
}
