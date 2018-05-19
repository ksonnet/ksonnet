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
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/ksonnet/ksonnet/pkg/app"
	"github.com/ksonnet/ksonnet/pkg/app/mocks"
	"github.com/ksonnet/ksonnet/pkg/util/test"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

func TestAddJPaths(t *testing.T) {
	withJsonnetPaths(func() {
		AddJPaths("/vendor")
		require.Equal(t, []string{"/vendor"}, componentJPaths)
	})
}

func TestAddExtVar(t *testing.T) {
	type args struct {
		key   string
		value string
	}
	testCases := []struct {
		name string
		args args
	}{
		{
			name: "add a key and value",
			args: args{
				key:   "key",
				value: "value",
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			withJsonnetPaths(func() {
				AddExtVar(tc.args.key, tc.args.value)
				require.Equal(t, tc.args.value, componentExtVars[tc.args.key])
			})
		})
	}
}

func TestAddExtVarFile(t *testing.T) {
	type args struct {
		key  string
		file string
	}
	testCases := []struct {
		name          string
		args          args
		expectedValue string
		stagePath     string
		isErr         bool
	}{
		{
			name: "add a key and value",
			args: args{
				key:  "key",
				file: "/app/value.txt",
			},
			expectedValue: "value",
			stagePath:     "/app/value.txt",
		},
		{
			name: "add a key and value",
			args: args{
				key:  "key",
				file: "/app/value.txt",
			},
			isErr: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			test.WithApp(t, "/app", func(a *mocks.App, fs afero.Fs) {
				withJsonnetPaths(func() {
					if tc.stagePath != "" {
						test.StageFile(t, fs, "value.txt", tc.stagePath)
					}
					err := AddExtVarFile(a, tc.args.key, tc.args.file)
					if tc.isErr {
						require.Error(t, err)
						return
					}
					require.NoError(t, err)
					require.Equal(t, tc.expectedValue, componentExtVars[tc.args.key])

				})
			})
		})
	}
}

func TestAddTlaVar(t *testing.T) {
	type args struct {
		key   string
		value string
	}
	testCases := []struct {
		name          string
		args          args
		expectedKey   string
		expectedValue string
	}{
		{
			name: "add a key and value",
			args: args{
				key:   "key",
				value: "value",
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			withJsonnetPaths(func() {
				AddTlaVar(tc.args.key, tc.args.value)
				require.Equal(t, tc.args.value, componentTlaVars[tc.args.key])
			})
		})
	}
}

func TestAddTlaVarFile(t *testing.T) {
	type args struct {
		key  string
		file string
	}
	testCases := []struct {
		name          string
		args          args
		expectedValue string
		stagePath     string
		isErr         bool
	}{
		{
			name: "add a key and value",
			args: args{
				key:  "key",
				file: "/app/value.txt",
			},
			expectedValue: "value",
			stagePath:     "/app/value.txt",
		},
		{
			name: "add a key and value",
			args: args{
				key:  "key",
				file: "/app/value.txt",
			},
			isErr: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			test.WithApp(t, "/app", func(a *mocks.App, fs afero.Fs) {
				withJsonnetPaths(func() {
					if tc.stagePath != "" {
						test.StageFile(t, fs, "value.txt", tc.stagePath)
					}
					err := AddTlaVarFile(a, tc.args.key, tc.args.file)
					if tc.isErr {
						require.Error(t, err)
						return
					}
					require.NoError(t, err)
					require.Equal(t, tc.expectedValue, componentTlaVars[tc.args.key])

				})
			})
		})
	}
}

func withJsonnetPaths(fn func()) {
	ogComponentJPaths := componentJPaths
	ogComponentExtVars := componentExtVars
	ogComponentTlaVars := componentTlaVars

	defer func() {
		componentJPaths = ogComponentJPaths
		componentExtVars = ogComponentExtVars
		componentTlaVars = ogComponentTlaVars
	}()

	fn()
}

func TestEvaluate(t *testing.T) {
	test.WithApp(t, "/app", func(a *mocks.App, fs afero.Fs) {
		envSpec := &app.EnvironmentSpec{
			Path: "default",
			Destination: &app.EnvironmentDestinationSpec{
				Server:    "http://example.com",
				Namespace: "default",
			},
		}
		a.On("Environment", "default").Return(envSpec, nil)

		test.StageFile(t, fs, "main.jsonnet", "/app/environments/default/main.jsonnet")

		components, err := ioutil.ReadFile(filepath.FromSlash("testdata/evaluate/components.jsonnet"))
		require.NoError(t, err)

		got, err := Evaluate(a, "default", string(components), "")
		require.NoError(t, err)

		test.AssertOutput(t, "evaluate/out.jsonnet", got)
	})
}

func TestMainFile(t *testing.T) {
	test.WithApp(t, "/app", func(a *mocks.App, fs afero.Fs) {
		envSpec := &app.EnvironmentSpec{}
		a.On("Environment", "default").Return(envSpec, nil)

		test.StageFile(t, fs, "main.jsonnet", "/app/environments/main.jsonnet")

		got, err := MainFile(a, "default")
		require.NoError(t, err)

		test.AssertOutput(t, "main.jsonnet", got)
	})
}

func Test_upgradeArray(t *testing.T) {
	snippet, err := ioutil.ReadFile(filepath.FromSlash("testdata/upgradeArray/in.jsonnet"))
	require.NoError(t, err)

	got, err := upgradeArray(string(snippet))
	require.NoError(t, err)

	test.AssertOutput(t, "upgradeArray/out.jsonnet", got)
}
