package metadata

import (
	"path"

	"github.com/ksonnet/ksonnet/metadata/app"
	"github.com/ksonnet/ksonnet/metadata/parts"
	"github.com/ksonnet/ksonnet/metadata/registry"
)

type mockRegistryManager struct {
	*app.RegistryRefSpec
	registryDir string
}

func newMockRegistryManager(name string) *mockRegistryManager {
	return &mockRegistryManager{
		registryDir: name,
		RegistryRefSpec: &app.RegistryRefSpec{
			Name: name,
		},
	}
}

func (m *mockRegistryManager) ResolveLibrarySpec(libID, libRefSpec string) (*parts.Spec, error) {
	return nil, nil
}

func (m *mockRegistryManager) RegistrySpecDir() string {
	return m.registryDir
}

func (m *mockRegistryManager) RegistrySpecFilePath() string {
	return path.Join(m.registryDir, "master.yaml")
}

func (m *mockRegistryManager) FetchRegistrySpec() (*registry.Spec, error) {
	registrySpec := registry.Spec{
		APIVersion: registry.DefaultApiVersion,
		Kind:       registry.DefaultKind,
	}

	return &registrySpec, nil
}

func (m *mockRegistryManager) MakeRegistryRefSpec() *app.RegistryRefSpec {
	return m.RegistryRefSpec
}

func (m *mockRegistryManager) ResolveLibrary(libID, libAlias, version string, onFile registry.ResolveFile, onDir registry.ResolveDirectory) (*parts.Spec, *app.LibraryRefSpec, error) {
	return nil, nil, nil
}
