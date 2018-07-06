// Copyright 2018 The ksonnet authors
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

package registry

import (
	"path/filepath"
	"testing"

	"github.com/ksonnet/ksonnet/pkg/app"
	amocks "github.com/ksonnet/ksonnet/pkg/app/mocks"
	"github.com/ksonnet/ksonnet/pkg/pkg"
	"github.com/ksonnet/ksonnet/pkg/util/test"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// We can't currently import registry/mocks due to a cycle.
// Implement simple mock registry.InstalledChecker.
type installedChecker struct{}

func (_m *installedChecker) IsInstalled(d pkg.Descriptor) (bool, error) {
	return false, nil
}

func Test_CacheDependency(t *testing.T) {
	withApp(t, func(a *amocks.App, fs afero.Fs) {
		a.On("VendorPath").Return("/app/vendor")

		test.StageDir(t, fs, "incubator", filepath.Join("/work", "incubator"))

		libraries := app.LibraryConfigs{}
		a.On("Libraries").Return(libraries, nil)

		registries := app.RegistryConfigs{
			"incubator": &app.RegistryConfig{
				Name:     "incubator",
				Protocol: string(ProtocolFilesystem),
				URI:      "/work/incubator",
			},
		}
		a.On("Registries").Return(registries, nil)

		libs := []*app.LibraryConfig{
			&app.LibraryConfig{
				Name:     "apache",
				Registry: "incubator",
			},
			&app.LibraryConfig{
				Name:     "nginx",
				Registry: "incubator",
				Version:  "", // TODO inject custom registry resolver so we can test versioned packages too
			},
		}
		for _, lib := range libs {
			a.On("UpdateLib", lib.Name, lib).Return(nil)

			var checker installedChecker
			d := pkg.Descriptor{Registry: lib.Registry, Name: lib.Name}

			_, err := CacheDependency(a, &checker, d, "")
			require.NoError(t, err)

			test.AssertExists(t, fs, filepath.Join(a.Root(), "vendor", lib.Registry, lib.Name, "parts.yaml"))
		}
	})
}

func Test_versionAndVendorRelPath(t *testing.T) {
	tests := []struct {
		name     string
		lib      app.LibraryConfig
		relPath  string
		expected string
	}{
		{
			name: "",
			lib: app.LibraryConfig{
				Name:     "nginx",
				Registry: "incubator",
				Version:  "1.2.3",
			},
			relPath:  "nginx/parts.yaml",
			expected: "/app/vendor/incubator/nginx@1.2.3/parts.yaml",
		},
		{
			name: "",
			lib: app.LibraryConfig{
				Name:     "nginx",
				Registry: "incubator",
				Version:  "",
			},
			relPath:  "nginx/parts.yaml",
			expected: "/app/vendor/incubator/nginx/parts.yaml",
		},
		{
			name: "",
			lib: app.LibraryConfig{
				Name:     "nginx",
				Registry: "incubator",
				Version:  "1.2.3",
			},
			relPath:  "nginx/longer/path",
			expected: "/app/vendor/incubator/nginx@1.2.3/longer/path",
		},
	}

	for _, tc := range tests {
		actual := versionAndVendorRelPath(&tc.lib, "/app/vendor", tc.relPath)
		assert.Equal(t, tc.expected, actual, tc.name)
	}
}
