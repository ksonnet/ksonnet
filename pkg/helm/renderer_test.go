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

package helm

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/ksonnet/ksonnet/pkg/app"
	amocks "github.com/ksonnet/ksonnet/pkg/app/mocks"
	"github.com/ksonnet/ksonnet/pkg/util/jsonnet"
	"github.com/ksonnet/ksonnet/pkg/util/test"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenderer_Render(t *testing.T) {
	cases := []struct {
		name    string
		version string
		isErr   bool
	}{
		{
			name:    "with version",
			version: "3.4.3",
		},
		{
			name: "without chart version",
		},
		{
			name:    "with invalid version",
			version: "invalid",
			isErr:   true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tmpDir, err := ioutil.TempDir("", "TestRenderer_Render")
			require.NoError(t, err)

			defer os.RemoveAll(tmpDir)

			fs := afero.NewOsFs()

			test.WithAppFs(t, tmpDir, fs, func(a *amocks.App, fs afero.Fs) {
				test.StageDir(t, fs, "redis", filepath.Join(a.Root(), "vendor", "helm-stable", "redis"))

				envConfig := &app.EnvironmentConfig{
					KubernetesVersion: "v1.10.3",
				}
				a.On("Environment", "default").Return(envConfig, nil)

				r := NewRenderer(a, "default")

				values := map[string]interface{}{}

				got, err := r.Render("helm-stable", "redis", tc.version, "componentName", values)
				if tc.isErr {
					require.Error(t, err)
					return
				}

				require.NoError(t, err)
				assert.NotEmpty(t, got)
			})
		})
	}
}

func TestRenderer_JsonnetNativeFunc(t *testing.T) {
	cases := []struct {
		name    string
		snippet string
		isErr   bool
	}{
		{
			name:    "with valid options",
			snippet: `std.prune(std.native("renderHelmChart")("helm-stable", "redis", "3.4.3", {}, "componentName"))`,
		},
		{
			name:    "with invalid options",
			snippet: `std.prune(std.native("renderHelmChart")("helm-stable"))`,
			isErr:   true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tmpDir, err := ioutil.TempDir("", "TestRenderer_Render")
			require.NoError(t, err)

			defer os.RemoveAll(tmpDir)

			fs := afero.NewOsFs()

			test.WithAppFs(t, tmpDir, fs, func(a *amocks.App, fs afero.Fs) {
				test.StageDir(t, fs, "redis", filepath.Join(a.Root(), "vendor", "helm-stable", "redis"))

				envConfig := &app.EnvironmentConfig{
					KubernetesVersion: "v1.10.3",
				}
				a.On("Environment", "default").Return(envConfig, nil)

				r := NewRenderer(a, "default")

				vm := jsonnet.NewVM()
				vm.AddFunctions(r.JsonnetNativeFunc())

				_, err := vm.EvaluateSnippet("snippet", tc.snippet)
				if tc.isErr {
					require.Error(t, err)
					return
				}
				require.NoError(t, err)
			})
		})
	}
}
