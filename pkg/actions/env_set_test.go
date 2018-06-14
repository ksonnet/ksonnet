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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnvSet_name(t *testing.T) {
	withApp(t, func(appMock *amocks.App) {
		envName := "old_env_name"
		newName := "new_env_name"

		in := map[string]interface{}{
			OptionApp:        appMock,
			OptionEnvName:    envName,
			OptionNewEnvName: newName,
		}

		a, err := NewEnvSet(in)
		require.NoError(t, err)

		a.envRenameFn = func(a app.App, from, to string, override bool) error {
			assert.Equal(t, envName, from)
			assert.Equal(t, newName, to)
			assert.False(t, override)

			return nil
		}

		spec := &app.EnvironmentSpec{
			Destination: &app.EnvironmentDestinationSpec{},
		}

		appMock.On("Environment", envName).Return(spec, nil)

		err = a.Run()
		require.NoError(t, err)
	})
}

func TestEnvSet_namespace(t *testing.T) {
	withApp(t, func(appMock *amocks.App) {
		envName := "env_name"
		oldNamespace := "old_namespace"
		namespace := "new_name_sapce"

		in := map[string]interface{}{
			OptionApp:       appMock,
			OptionEnvName:   envName,
			OptionNamespace: namespace,
		}
		a, err := NewEnvSet(in)
		require.NoError(t, err)

		spec := &app.EnvironmentSpec{
			Destination: &app.EnvironmentDestinationSpec{
				Namespace: oldNamespace,
			},
		}

		updatedSpec := &app.EnvironmentSpec{
			Destination: &app.EnvironmentDestinationSpec{
				Namespace: namespace,
			},
		}

		appMock.On("Environment", envName).Return(spec, nil)
		appMock.On("AddEnvironment", envName, "", updatedSpec, false).Return(nil)

		err = a.Run()
		require.NoError(t, err)
	})
}

func TestEnvSet_server(t *testing.T) {
	withApp(t, func(appMock *amocks.App) {
		envName := "env_name"
		oldServer := "old_server"
		server := "new_server"

		in := map[string]interface{}{
			OptionApp:     appMock,
			OptionEnvName: envName,
			OptionServer:  server,
		}
		a, err := NewEnvSet(in)
		require.NoError(t, err)

		spec := &app.EnvironmentSpec{
			Destination: &app.EnvironmentDestinationSpec{
				Server: oldServer,
			},
		}

		updatedSpec := &app.EnvironmentSpec{
			Destination: &app.EnvironmentDestinationSpec{
				Server: server,
			},
		}

		appMock.On("Environment", envName).Return(spec, nil)
		appMock.On("AddEnvironment", envName, "", updatedSpec, false).Return(nil)

		err = a.Run()
		require.NoError(t, err)
	})
}
func TestEnvSet_all(t *testing.T) {
	withApp(t, func(appMock *amocks.App) {
		envName := "old_env_name"
		newName := "new_env_name"
		oldNamespace := "old_namespace"
		namespace := "new_name_sapce"
		oldServer := "old_server"
		server := "new_server"

		in := map[string]interface{}{
			OptionApp:        appMock,
			OptionEnvName:    envName,
			OptionNewEnvName: newName,
			OptionNamespace:  namespace,
			OptionServer:     server,
		}

		a, err := NewEnvSet(in)
		require.NoError(t, err)

		a.envRenameFn = func(a app.App, from, to string, override bool) error {
			assert.Equal(t, envName, from)
			assert.Equal(t, newName, to)
			assert.False(t, override)

			return nil
		}

		spec := &app.EnvironmentSpec{
			Destination: &app.EnvironmentDestinationSpec{
				Namespace: oldNamespace,
				Server:    oldServer,
			},
		}

		updatedSpec := &app.EnvironmentSpec{
			Destination: &app.EnvironmentDestinationSpec{
				Namespace: namespace,
				Server:    server,
			},
		}

		appMock.On("Environment", envName).Return(spec, nil)
		appMock.On("AddEnvironment", newName, "", updatedSpec, false).Return(nil)

		err = a.Run()
		require.NoError(t, err)
	})
}

func TestEnvSet_requires_app(t *testing.T) {
	in := make(map[string]interface{})
	_, err := NewEnvSet(in)
	require.Error(t, err)
}
