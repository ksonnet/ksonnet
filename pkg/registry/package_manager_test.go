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
	"testing"

	"github.com/ksonnet/ksonnet/pkg/app"
	amocks "github.com/ksonnet/ksonnet/pkg/app/mocks"
	"github.com/ksonnet/ksonnet/pkg/parts"
	"github.com/ksonnet/ksonnet/pkg/util/test"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_packageManager_Find(t *testing.T) {
	test.WithApp(t, "/app", func(a *amocks.App, fs afero.Fs) {

		test.StageDir(t, fs, "incubator/apache", "/work/apache")
		test.StageDir(t, fs, "incubator/apache", "/app/vendor/incubator/apache")

		a.On("VendorPath").Return("/app/vendor")

		registries := app.RegistryRefSpecs{
			"incubator": &app.RegistryRefSpec{
				Protocol: "fs",
				URI:      "/work",
			},
		}

		a.On("Registries").Return(registries, nil)

		libraries := app.LibraryRefSpecs{
			"apache": &app.LibraryRefSpec{},
		}

		a.On("Libraries").Return(libraries, nil)

		pm := NewPackageManager(a)

		p, err := pm.Find("incubator/apache")
		require.NoError(t, err)

		require.Equal(t, "apache", p.Name())
	})
}

func Test_packageManager_Packages(t *testing.T) {
	test.WithApp(t, "/app", func(a *amocks.App, fs afero.Fs) {

		test.StageDir(t, fs, "incubator/apache", "/work/apache")
		test.StageDir(t, fs, "incubator/apache", "/app/vendor/incubator/apache")

		a.On("VendorPath").Return("/app/vendor")

		registries := app.RegistryRefSpecs{
			"incubator": &app.RegistryRefSpec{
				Protocol: "fs",
				URI:      "/work",
			},
		}

		a.On("Registries").Return(registries, nil)

		libraries := app.LibraryRefSpecs{
			"apache": &app.LibraryRefSpec{
				Registry: "incubator",
			},
		}

		a.On("Libraries").Return(libraries, nil)

		pm := NewPackageManager(a)

		packages, err := pm.Packages()
		require.NoError(t, err)

		require.Len(t, packages, 1)
		p := packages[0]

		require.Equal(t, "apache", p.Name())
	})
}

func Test_packageManager_Prototypes(t *testing.T) {
	test.WithApp(t, "/app", func(a *amocks.App, fs afero.Fs) {

		test.StageDir(t, fs, "incubator/apache", "/work/apache")
		test.StageDir(t, fs, "incubator/apache", "/app/vendor/incubator/apache")

		a.On("VendorPath").Return("/app/vendor")

		registries := app.RegistryRefSpecs{
			"incubator": &app.RegistryRefSpec{
				Protocol: "fs",
				URI:      "/work",
			},
		}

		a.On("Registries").Return(registries, nil)

		libraries := app.LibraryRefSpecs{
			"apache": &app.LibraryRefSpec{
				Registry: "incubator",
			},
		}

		a.On("Libraries").Return(libraries, nil)

		pm := NewPackageManager(a)

		protos, err := pm.Prototypes()
		require.NoError(t, err)

		require.Len(t, protos, 1)
	})
}

func Test_remotePackage(t *testing.T) {
	rp := &remotePackage{
		registryName: "registry-name",
		partConfig: &parts.Spec{
			Name:        "pkg-name",
			Description: "description",
		},
	}

	i, err := rp.IsInstalled()
	require.NoError(t, err)

	protos, err := rp.Prototypes()
	require.NoError(t, err)

	assert.Equal(t, "pkg-name", rp.Name())
	assert.Equal(t, "registry-name", rp.RegistryName())
	assert.Equal(t, "description", rp.Description())
	assert.False(t, i)
	assert.Empty(t, protos)
}
