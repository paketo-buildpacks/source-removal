package nosource

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/cloudfoundry/packit"
)

func Build() packit.BuildFunc {
	return func(context packit.BuildContext) (packit.BuildResult, error) {
		files, err := ioutil.ReadDir(context.WorkingDir)
		if err != nil {
			return packit.BuildResult{}, err
		}

		for _, f := range files {
			err = os.RemoveAll(filepath.Join(context.WorkingDir, f.Name()))
			if err != nil {
				return packit.BuildResult{}, err
			}
		}
		return packit.BuildResult{}, nil
	}
}
