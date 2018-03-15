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

package env

import (
	"testing"

	"github.com/ksonnet/ksonnet/metadata/app"
	"github.com/ksonnet/ksonnet/metadata/app/mocks"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

func TestList(t *testing.T) {
	withEnv(t, func(appMock *mocks.App, fs afero.Fs) {
		specEnvs := app.EnvironmentSpecs{
			"default": &app.EnvironmentSpec{
				Path: "default",
				Destination: &app.EnvironmentDestinationSpec{
					Namespace: "default",
					Server:    "http://example.com",
				},
				KubernetesVersion: "v1.8.7",
			},
		}
		appMock.On("Environments").Return(specEnvs, nil)

		envs, err := List(appMock)
		require.NoError(t, err)

		expected := map[string]Env{
			"default": Env{
				KubernetesVersion: "v1.8.7",
				Name:              "default",
				Destination:       NewDestination("http://example.com", "default"),
			},
		}

		require.Equal(t, expected, envs)
	})
}
