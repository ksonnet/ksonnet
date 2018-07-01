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

package params

import (
	"path/filepath"
	"testing"

	"github.com/ksonnet/ksonnet/pkg/app"
	"github.com/ksonnet/ksonnet/pkg/app/mocks"
	"github.com/ksonnet/ksonnet/pkg/util/test"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEvaluateEnv(t *testing.T) {
	test.WithApp(t, "/app", func(a *mocks.App, fs afero.Fs) {
		envConfig := &app.EnvironmentConfig{
			Destination: &app.EnvironmentDestinationSpec{
				Namespace: "default",
				Server:    "http://example.com",
			},
		}
		a.On("Environment", "default").Return(envConfig, nil)

		sourcePath := "/app/environments/default/params.libsonnet"
		paramsStr := test.ReadTestData(t, filepath.Join("evaluate_env", "component_params.libsonnet"))
		envName := "default"
		moduleName := "app.project-1"

		test.StageFile(t, fs, filepath.Join("evaluate_env", "env_params.libsonnet"), sourcePath)

		got, err := EvaluateEnv(a, sourcePath, paramsStr, envName, moduleName)
		require.NoError(t, err)

		expected := test.ReadTestData(t, filepath.Join("evaluate_env", "expected.libsonnet"))

		assert.Equal(t, expected, got)
	})
}
