package sourceremoval

import (
	"os"

	"github.com/paketo-buildpacks/source-removal/logic"

	"github.com/paketo-buildpacks/packit/v2"
)

func Build() packit.BuildFunc {
	return func(context packit.BuildContext) (packit.BuildResult, error) {
		if includeVal, ok := os.LookupEnv("BP_INCLUDE_FILES"); ok {
			err := logic.Include(context.WorkingDir, includeVal)
			if err != nil {
				return packit.BuildResult{}, err
			}
		}

		if excludeVal, ok := os.LookupEnv("BP_EXCLUDE_FILES"); ok {
			err := logic.Exclude(context.WorkingDir, excludeVal)
			if err != nil {
				return packit.BuildResult{}, err
			}
		}

		return packit.BuildResult{}, nil
	}
}
