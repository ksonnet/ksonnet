// Copyright 2018 The kubecfg authors
//
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.

package metadata

import (
	"path"

	"github.com/ksonnet/ksonnet/metadata/app"
	"github.com/ksonnet/ksonnet/pkg/parts"
	"github.com/ksonnet/ksonnet/pkg/registry"
)

//
// Mock registry manager for end-to-end tests.
//

type mockRegistryManager struct {
	*app.RegistryRefSpec
	registryDir string
}

var _ registry.Registry = (*mockRegistryManager)(nil)

func newMockRegistryManager(name string) *mockRegistryManager {
	return &mockRegistryManager{
		registryDir: name,
		RegistryRefSpec: &app.RegistryRefSpec{
			Name: name,
		},
	}
}

func (m *mockRegistryManager) Name() string {
	return m.registryDir
}

func (m *mockRegistryManager) Protocol() string {
	return registry.ProtocolGitHub
}

func (m *mockRegistryManager) URI() string {
	return "github.com/foo/bar"
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
		APIVersion: registry.DefaultAPIVersion,
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
