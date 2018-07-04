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
	"bytes"
	"path/filepath"
	"testing"

	"github.com/ksonnet/ksonnet/pkg/app"
	amocks "github.com/ksonnet/ksonnet/pkg/app/mocks"
	"github.com/ksonnet/ksonnet/pkg/util/test"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestEnvList(t *testing.T) {

	setupValidApp := func(appMock *amocks.App) {
		defaultEnv := &app.EnvironmentConfig{
			KubernetesVersion: "v1.7.0",
			Destination: &app.EnvironmentDestinationSpec{
				Namespace: "default",
				Server:    "http://example.com",
			},
		}

		prodEnv := &app.EnvironmentConfig{
			KubernetesVersion: "v1.7.0",
			Destination: &app.EnvironmentDestinationSpec{
				Namespace: "prod",
				Server:    "http://example.com",
			},
		}

		envs := app.EnvironmentConfigs{
			"default": defaultEnv,
			"prod":    prodEnv,
		}

		appMock.On("Environments").Return(envs, nil)
	}

	envListFail := func(appMock *amocks.App) {
		appMock.On("Environments").Return(nil, errors.New("failed"))
	}

	cases := []struct {
		name         string
		initApp      func(*amocks.App)
		outputType   string
		expectedFile string
		isErr        bool
	}{
		{
			name:         "table output",
			initApp:      setupValidApp,
			outputType:   "table",
			expectedFile: filepath.Join("env", "list", "output.txt"),
		},
		{
			name:         "no format specified",
			initApp:      setupValidApp,
			expectedFile: filepath.Join("env", "list", "output.txt"),
		},
		{
			name:         "json output",
			initApp:      setupValidApp,
			outputType:   "json",
			expectedFile: filepath.Join("env", "list", "output.json"),
		},
		{
			name:       "invalid output format",
			initApp:    setupValidApp,
			outputType: "invalid",
			isErr:      true,
		},
		{
			name:    "environment list failed",
			initApp: envListFail,
			isErr:   true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			withApp(t, func(appMock *amocks.App) {
				tc.initApp(appMock)
				in := map[string]interface{}{
					OptionApp:    appMock,
					OptionOutput: tc.outputType,
				}

				a, err := NewEnvList(in)
				require.NoError(t, err)

				var buf bytes.Buffer
				a.out = &buf

				err = a.Run()
				if tc.isErr {
					require.Error(t, err)
					return
				}

				require.NoError(t, err)
				test.AssertOutput(t, tc.expectedFile, buf.String())
			})
		})
	}
}

func TestEnvList_requires_app(t *testing.T) {
	in := make(map[string]interface{})
	_, err := NewEnvList(in)
	require.Error(t, err)
}
