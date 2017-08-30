package kubecfg

import (
	"fmt"

	"github.com/ksonnet/kubecfg/metadata"
)

func GetFiles(wd metadata.AbsPath, env *string, files []string) ([]string, error) {
	envPresent := env != nil
	filesPresent := len(files) > 0

	// This is equivalent to: `if !xor(envPresent, filesPresent) {`
	if envPresent && filesPresent {
		return nil, fmt.Errorf("Either an environment name or a file list is required, but not both")
	} else if !envPresent && !filesPresent {
		return nil, fmt.Errorf("Must specify either an environment or a file list")
	}

	if envPresent {
		manager, err := metadata.Find(wd)
		if err != nil {
			return nil, err
		}

		files, err = manager.ComponentPaths()
		if err != nil {
			return nil, err
		}
	}

	return files, nil
}
