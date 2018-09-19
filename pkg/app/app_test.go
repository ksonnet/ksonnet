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

package app

import (
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeTransport struct {
	resp *http.Response
	err  error
}

func (f *fakeTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return f.resp, f.err
}

// fakeHTTPClient returns an http.Client that will respond with the predefined response and error.
func fakeHTTPClient(resp *http.Response, err error) *http.Client {
	c := &http.Client{
		Transport: &fakeTransport{
			resp: resp,
			err:  err,
		},
	}
	return c
}

func Test_findRoot(t *testing.T) {
	fs := afero.NewMemMapFs()
	stageFile(t, fs, "app010_app.yaml", "/app/app.yaml")

	dirs := []string{
		"/app/components",
		"/invalid",
	}

	for _, dir := range dirs {
		err := fs.MkdirAll(dir, DefaultFilePermissions)
		require.NoError(t, err)
	}

	cases := []struct {
		name     string
		expected string
		isErr    bool
	}{
		{
			name:     "/app",
			expected: "/app",
		},
		{
			name:     "/app/components",
			expected: "/app",
		},
		{
			name:  "/invalid",
			isErr: true,
		},
		{
			name:  "/missing",
			isErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root, err := FindRoot(fs, tc.name)
			if tc.isErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.expected, root)
		})
	}

}

func TestApp_AddEnvironment(t *testing.T) {
	withAppFs(t, "app010_app.yaml", func(app *baseApp) {
		envs, err := app.Environments()
		require.NoError(t, err)

		envLen := len(envs)

		newEnv := &EnvironmentConfig{
			Name: "us-west/qa",
			Destination: &EnvironmentDestinationSpec{
				Namespace: "some-namespace",
				Server:    "http://example.com",
			},
			Path: "us-west/qa",
		}

		k8sSpecFlag := "version:v1.8.7"
		err = app.AddEnvironment(newEnv, k8sSpecFlag, false)
		require.NoError(t, err)

		envs, err = app.Environments()
		require.NoError(t, err)
		require.Len(t, envs, envLen+1)

		env, err := app.Environment("us-west/qa")
		require.NoError(t, err)
		require.Equal(t, "v1.8.7", env.KubernetesVersion)
	})
}

func Test_AddEnvironment_empty_spec_flag(t *testing.T) {
	withAppFs(t, "app010_app.yaml", func(app *baseApp) {
		envs, err := app.Environments()
		require.NoError(t, err)

		envLen := len(envs)

		env, err := app.Environment("default")
		require.NoError(t, err)

		env.Destination.Namespace = "updated"

		err = app.AddEnvironment(env, "", false)
		require.NoError(t, err)

		envs, err = app.Environments()
		require.NoError(t, err)
		require.Len(t, envs, envLen)

		env, err = app.Environment("default")
		require.NoError(t, err)
		require.Equal(t, "v1.7.0", env.KubernetesVersion)
		require.Equal(t, "updated", env.Destination.Namespace)
	})
}

func TestApp_Environments(t *testing.T) {
	withAppFs(t, "app010_app.yaml", func(app *baseApp) {
		expected := EnvironmentConfigs{
			"default": &EnvironmentConfig{
				Name: "default",
				Destination: &EnvironmentDestinationSpec{
					Namespace: "some-namespace",
					Server:    "http://example.com",
				},
				KubernetesVersion: "v1.7.0",
				Libraries:         LibraryConfigs{},
				Path:              "default",
				Targets:           []string{},
			},
			"us-east/test": &EnvironmentConfig{
				Name: "us-east/test",
				Destination: &EnvironmentDestinationSpec{
					Namespace: "some-namespace",
					Server:    "http://example.com",
				},
				KubernetesVersion: "v1.7.0",
				Libraries:         LibraryConfigs{},
				Path:              "us-east/test",
				Targets:           []string{},
			},
			"us-west/test": &EnvironmentConfig{
				Name: "us-west/test",
				Destination: &EnvironmentDestinationSpec{
					Namespace: "some-namespace",
					Server:    "http://example.com",
				},
				KubernetesVersion: "v1.7.0",
				Libraries:         LibraryConfigs{},
				Path:              "us-west/test",
				Targets:           []string{},
			},
			"us-west/prod": &EnvironmentConfig{
				Name: "us-west/prod",
				Destination: &EnvironmentDestinationSpec{
					Namespace: "some-namespace",
					Server:    "http://example.com",
				},
				KubernetesVersion: "v1.7.0",
				Libraries:         LibraryConfigs{},
				Path:              "us-west/prod",
				Targets:           []string{},
			},
		}
		envs, err := app.Environments()
		require.NoError(t, err)

		require.Equal(t, expected, envs)
	})
}

func TestApp_Environment(t *testing.T) {
	cases := []struct {
		name    string
		envName string
		isErr   bool
	}{
		{
			name:    "existing env",
			envName: "us-east/test",
		},
		{
			name:    "invalid env",
			envName: "missing",
			isErr:   true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			withAppFs(t, "app010_app.yaml", func(app *baseApp) {
				spec, err := app.Environment(tc.envName)
				if tc.isErr {
					require.Error(t, err)
				} else {
					require.NoError(t, err)
					require.Equal(t, tc.envName, spec.Path)
				}
			})
		})
	}
}

func Test_App_Environment_returns_copy(t *testing.T) {
	withAppFs(t, "app010_app.yaml", func(app *baseApp) {
		e1, err := app.Environment("default")
		require.NoError(t, err)

		e1.KubernetesVersion = "v9.9.9"

		e2, err := app.Environment("default")
		require.NoError(t, err)
		assert.False(t, e1 == e2, "expected new pointer")
		assert.Equal(t, "v9.9.9", e1.KubernetesVersion)
		assert.Equal(t, "v1.7.0", e2.KubernetesVersion)
	})
}

func TestApp_LibPath(t *testing.T) {
	withAppFs(t, "app010_app.yaml", func(app *baseApp) {
		app.libUpdater = fakeLibUpdater(func(string, string) (string, error) {
			return "v1.8.7", nil
		})

		specReader, err := os.Open("../cluster/testdata/swagger.json")
		if err != nil {
			require.NoError(t, err, "opening fixture: swagger.json")
			return
		}
		defer specReader.Close()
		app.httpClient = fakeHTTPClient(
			&http.Response{
				StatusCode: 200,
				Body:       specReader,
			}, nil)
		path, err := app.LibPath("default")
		require.NoError(t, err)

		expected := filepath.Join("/", "lib", "ksonnet-lib", "v1.7.0")
		require.Equal(t, expected, path)
	})
}

func TestApp_RemoveEnvironment(t *testing.T) {
	withAppFs(t, "app010_app.yaml", func(app *baseApp) {
		_, err := app.Environment("default")
		require.NoError(t, err)

		err = app.RemoveEnvironment("default", false)
		require.NoError(t, err)

		_, err = app.Environment("default")
		require.Error(t, err)

		err = app.RemoveEnvironment("invalid", false)
		require.Error(t, err)
	})
}

func TestApp_RenameEnvironment(t *testing.T) {
	cases := []struct {
		name           string
		from           string
		to             string
		shouldExist    []string
		shouldNotExist []string
	}{
		{
			name: "rename",
			from: "default",
			to:   "renamed",
			shouldExist: []string{
				"/environments/renamed/main.jsonnet",
			},
			shouldNotExist: []string{
				"/environments/default",
			},
		},
		{
			name: "rename to nested",
			from: "default",
			to:   "default/nested",
			shouldExist: []string{
				"/environments/default/nested/main.jsonnet",
			},
			shouldNotExist: []string{
				"/environments/default/main.jsonnet",
			},
		},
		{
			name: "un-nest",
			from: "us-east/test",
			to:   "us-east",
			shouldExist: []string{
				"/environments/us-east/main.jsonnet",
			},
			shouldNotExist: []string{
				"/environments/us-east/test",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			withAppFs(t, "app010_app.yaml", func(app *baseApp) {
				err := app.RenameEnvironment(tc.from, tc.to, false)
				require.NoError(t, err)

				for _, p := range tc.shouldExist {
					checkExist(t, app.Fs(), p)
				}

				for _, p := range tc.shouldNotExist {
					checkNotExist(t, app.Fs(), p)
				}

				_, err = app.Environment(tc.from)
				assert.Error(t, err)

				_, err = app.Environment(tc.to)
				assert.NoError(t, err)
			})
		})
	}
}

func TestApp_UpdateTargets(t *testing.T) {
	withAppFs(t, "app010_app.yaml", func(app *baseApp) {
		err := app.UpdateTargets("default", []string{"foo"}, false)
		require.NoError(t, err)

		e, err := app.Environment("default")
		require.NoError(t, err)

		expected := []string{"foo"}
		require.Equal(t, expected, e.Targets)
	})
}

type fakeLibUpdater func(k8sSpecFlag string, libPath string) (string, error)

func (f fakeLibUpdater) UpdateKSLib(k8sSpecFlag string, libPath string) (string, error) {
	return f(k8sSpecFlag, libPath)
}

func withAppFs(t *testing.T, appName string, fn func(app *baseApp)) {
	dir, err := ioutil.TempDir("", "")
	require.NoError(t, err)

	defer os.RemoveAll(dir)

	fs := afero.NewBasePathFs(afero.NewOsFs(), dir)

	envDirs := []string{
		"default",
		"us-east/test",
		"us-west/test",
		"us-west/prod",
	}

	for _, dir := range envDirs {
		path := filepath.Join("/environments", dir)
		err = fs.MkdirAll(path, DefaultFolderPermissions)
		require.NoError(t, err)

		swaggerPath := filepath.Join(path, "main.jsonnet")
		stageFile(t, fs, "main.jsonnet", swaggerPath)
	}

	stageFile(t, fs, appName, "/app.yaml")

	optLibUpdater := OptLibUpdater(
		fakeLibUpdater(func(string, string) (string, error) {
			return "v1.8.7", nil
		}),
	)

	app := NewBaseApp(fs, "/", nil, optLibUpdater)

	fn(app)
}

func TestApp_Load(t *testing.T) {
	fs := afero.NewMemMapFs()

	expectedEnvs := EnvironmentConfigs{
		"default": &EnvironmentConfig{
			Name: "default",
			Destination: &EnvironmentDestinationSpec{
				Namespace: "some-namespace",
				Server:    "http://example.com",
			},
			KubernetesVersion: "v1.7.0",
			Path:              "default",
		},
		"us-east/test": &EnvironmentConfig{
			Name: "us-east/test",
			Destination: &EnvironmentDestinationSpec{
				Namespace: "some-namespace",
				Server:    "http://example.com",
			},
			KubernetesVersion: "v1.7.0",
			Path:              "us-east/test",
		},
		"us-west/prod": &EnvironmentConfig{
			Name: "us-west/prod",
			Destination: &EnvironmentDestinationSpec{
				Namespace: "some-namespace",
				Server:    "http://example.com",
			},
			KubernetesVersion: "v1.7.0",
			Path:              "us-west/prod",
		},
		"us-west/test": &EnvironmentConfig{
			Name: "us-west/test",
			Destination: &EnvironmentDestinationSpec{
				Namespace: "some-namespace",
				Server:    "http://example.com",
			},
			KubernetesVersion: "v1.7.0",
			Path:              "us-west/test",
		},
	}

	stageFile(t, fs, "app030_app.yaml", "/app.yaml")

	a, err := Load(fs, nil, "/")
	require.NoError(t, err)

	envs, err := a.Environments()
	require.NoError(t, err, "loading environments")

	assert.Equal(t, expectedEnvs, envs, "unexpected app content")
}

// Tests that a new app is initialized when an app.yaml does not yet exist
func TestApp_Load_no_cfg(t *testing.T) {
	fs := afero.NewMemMapFs()

	a, err := Load(fs, nil, "/")
	require.NoError(t, err)
	require.NotNil(t, a)

	_, err = fs.Stat("/app.yaml")
	require.True(t, os.IsNotExist(err), "app.yaml should not exist")
}
