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

package pkg

import (
	"testing"

	amocks "github.com/ksonnet/ksonnet/pkg/app/mocks"
	"github.com/ksonnet/ksonnet/pkg/util/test"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func withLocalPackage(t *testing.T, fn func(a *amocks.App, fs afero.Fs)) {
	test.WithApp(t, "/app", func(a *amocks.App, fs afero.Fs) {
		a.On("VendorPath").Return("/app/vendor")

		test.StageDir(t, fs, "incubator/apache", "/app/vendor/incubator/apache")

		fn(a, fs)
	})
}

func TestLocal_Name(t *testing.T) {
	withLocalPackage(t, func(a *amocks.App, fs afero.Fs) {
		l, err := NewLocal(a, "apache", "incubator", nil)
		require.NoError(t, err)

		require.Equal(t, "apache", l.Name())
	})
}

func TestLocal_RegistryName(t *testing.T) {
	withLocalPackage(t, func(a *amocks.App, fs afero.Fs) {
		l, err := NewLocal(a, "apache", "incubator", nil)
		require.NoError(t, err)

		require.Equal(t, "incubator", l.RegistryName())
	})
}

func TestLocal_IsInstalled(t *testing.T) {
	withLocalPackage(t, func(a *amocks.App, fs afero.Fs) {
		ic := &fakeInstallChecker{
			isInstalled: true,
		}

		l, err := NewLocal(a, "apache", "incubator", ic)
		require.NoError(t, err)

		i, err := l.IsInstalled()
		assert.NoError(t, err)
		assert.True(t, i)
	})
}

func TestLocal_Description(t *testing.T) {
	withLocalPackage(t, func(a *amocks.App, fs afero.Fs) {
		l, err := NewLocal(a, "apache", "incubator", nil)
		require.NoError(t, err)

		require.Equal(t, "part description", l.Description())
	})
}

func TestLocal_Prototypes(t *testing.T) {
	withLocalPackage(t, func(a *amocks.App, fs afero.Fs) {
		l, err := NewLocal(a, "apache", "incubator", nil)
		require.NoError(t, err)

		prototypes, err := l.Prototypes()
		require.NoError(t, err)

		require.Len(t, prototypes, 1)
		proto := prototypes[0]
		require.Equal(t, "io.ksonnet.pkg.apache-simple", proto.Name)
	})
}
