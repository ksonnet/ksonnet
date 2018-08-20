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

package actions

import (
	"testing"

	"github.com/ksonnet/ksonnet/pkg/app"
	amocks "github.com/ksonnet/ksonnet/pkg/app/mocks"
	"github.com/ksonnet/ksonnet/pkg/pkg"
	"github.com/ksonnet/ksonnet/pkg/registry"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPkgInstall(t *testing.T) {
	withApp(t, func(appMock *amocks.App) {
		libName := "incubator/apache"
		customName := "customName"

		in := map[string]interface{}{
			OptionApp:           appMock,
			OptionPkgName:       libName,
			OptionName:          customName,
			OptionForce:         false,
			OptionTLSSkipVerify: false,
		}

		a, err := NewPkgInstall(in)
		require.NoError(t, err)

		newLibCfg := &app.LibraryConfig{
			Registry: "incubator",
			Name:     "apache",
		}
		expectedD := pkg.Descriptor{
			Registry: "incubator",
			Name:     "apache",
		}

		var cacherCalled bool
		fakeCacher := func(a app.App, checker registry.InstalledChecker, d pkg.Descriptor, cn string, force bool) (*app.LibraryConfig, error) {
			cacherCalled = true
			require.Equal(t, expectedD, d)
			require.Equal(t, "customName", cn)
			return newLibCfg, nil
		}

		var updaterCalled bool
		fakeUpdater := func(name string, env string, spec *app.LibraryConfig) (*app.LibraryConfig, error) {
			updaterCalled = true
			assert.Equal(t, newLibCfg.Name, name, "unexpected library name")
			assert.Equal(t, a.envName, env, "unexpected environment name")
			assert.Equal(t, newLibCfg, spec, "unexpected library configuration object")
			if spec != nil {
				assert.Equal(t, expectedD.Name, spec.Name, "unexpected library name in configuration object")
			}
			return nil, nil
		}

		a.libCacherFn = fakeCacher
		a.libUpdateFn = fakeUpdater

		libraries := app.LibraryConfigs{}
		appMock.On("Libraries").Return(libraries, nil)

		registries := app.RegistryConfigs{
			"incubator": &app.RegistryConfig{
				Protocol: string(registry.ProtocolFilesystem),
				URI:      "file:///tmp",
			},
		}
		appMock.On("Registries").Return(registries, nil)

		err = a.Run()
		require.NoError(t, err)
		assert.True(t, cacherCalled, "dependency cacher not called")
		assert.True(t, updaterCalled, "library reference updater not called")
	})
}

func TestPkgInstall_requires_app(t *testing.T) {
	in := make(map[string]interface{})
	_, err := NewPkgInstall(in)
	require.Error(t, err)
}

func TestPkgInstall_invalid_env(t *testing.T) {
	withApp(t, func(appMock *amocks.App) {
		libName := "incubator/apache"
		customName := "customName"

		in := map[string]interface{}{
			OptionApp:           appMock,
			OptionPkgName:       libName,
			OptionName:          customName,
			OptionEnvName:       "invalid-env",
			OptionForce:         false,
			OptionTLSSkipVerify: false,
		}

		a, err := NewPkgInstall(in)
		require.NoError(t, err)

		var cacherCalled bool
		fakeCacher := func(a app.App, checker registry.InstalledChecker, d pkg.Descriptor, cn string, force bool) (*app.LibraryConfig, error) {
			cacherCalled = true
			return nil, errors.New("not implemented")
		}

		var updaterCalled bool
		fakeUpdater := func(name string, env string, spec *app.LibraryConfig) (*app.LibraryConfig, error) {
			updaterCalled = true
			return nil, errors.New("not implemented")
		}

		a.libCacherFn = fakeCacher
		a.libUpdateFn = fakeUpdater
		a.envCheckerFn = func(string) (bool, error) {
			return false, nil
		}

		err = a.Run()
		require.Error(t, err)
		assert.False(t, cacherCalled, "dependency cacher called unexpectedly")
		assert.False(t, updaterCalled, "library reference updater called unexpectedly")
	})
}
