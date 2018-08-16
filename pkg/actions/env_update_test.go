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
	"net/http"
	"testing"

	"github.com/ksonnet/ksonnet/pkg/app"
	amocks "github.com/ksonnet/ksonnet/pkg/app/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnvUpdate(t *testing.T) {
	withApp(t, func(appMock *amocks.App) {

		envSpec := &app.EnvironmentConfig{
			KubernetesVersion: "v1.8.9",
		}
		appMock.On("Environment", "envName").Return(envSpec, nil)
		appMock.On("LibPath", "envName").Return("/app/lib/v1.8.9", nil)

		in := map[string]interface{}{
			OptionApp:     appMock,
			OptionEnvName: "envName",
		}

		a, err := newEnvUpdate(in)
		require.NoError(t, err)

		a.genLibFn = func(_ app.App, k8sSpecFlag, libPath string, httpClient *http.Client) error {
			assert.Equal(t, "version:v1.8.9", k8sSpecFlag)
			assert.Equal(t, "/app/lib/v1.8.9", libPath)
			return nil
		}

		err = a.run()
		require.NoError(t, err)
	})
}

func TestEnvUpdate_requires_app(t *testing.T) {
	in := make(map[string]interface{})
	_, err := newEnvUpdate(in)
	require.Error(t, err)
}
