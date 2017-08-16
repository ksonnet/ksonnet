package metadata

import (
	"github.com/spf13/afero"
)

type clusterSpecFile struct {
	specPath AbsPath
}

func (cs *clusterSpecFile) data() ([]byte, error) {
	return afero.ReadFile(appFS, string(cs.specPath))
}

func (cs *clusterSpecFile) resource() string {
	return string(cs.specPath)
}

type clusterSpecLive struct {
	apiServerURL string
}

func (cs *clusterSpecLive) data() ([]byte, error) {
	// TODO: Implement getting spec from path, k8sVersion, and URL.
	panic("Not implemented")
}

func (cs *clusterSpecLive) resource() string {
	return string(cs.apiServerURL)
}

type clusterSpecVersion struct {
	k8sVersion string
}

func (cs *clusterSpecVersion) data() ([]byte, error) {
	// TODO: Implement getting spec from path, k8sVersion, and URL.
	panic("Not implemented")
}

func (cs *clusterSpecVersion) resource() string {
	return string(cs.k8sVersion)
}
