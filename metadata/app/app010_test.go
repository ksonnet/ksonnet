package app

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

func TestApp0101_Environments(t *testing.T) {
	withApp010Fs(t, "app010_app.yaml", func(fs afero.Fs) {
		app, err := NewApp010(fs, "/")
		require.NoError(t, err)

		expected := EnvironmentSpecs{
			"default": &EnvironmentSpec{
				Destination: &EnvironmentDestinationSpec{
					Namespace: "some-namespace",
					Server:    "http://example.com",
				},
				KubernetesVersion: "v1.7.0",
				Path:              "default",
			},
			"us-east/test": &EnvironmentSpec{
				Destination: &EnvironmentDestinationSpec{
					Namespace: "some-namespace",
					Server:    "http://example.com",
				},
				Path: "us-east/test",
			},
			"us-west/test": &EnvironmentSpec{
				Destination: &EnvironmentDestinationSpec{
					Namespace: "some-namespace",
					Server:    "http://example.com",
				},
				Path: "us-west/test",
			},
			"us-west/prod": &EnvironmentSpec{
				Destination: &EnvironmentDestinationSpec{
					Namespace: "some-namespace",
					Server:    "http://example.com",
				},
				Path: "us-west/prod",
			},
		}
		envs, err := app.Environments()
		require.NoError(t, err)

		require.Equal(t, expected, envs)
	})
}

func TestApp010_Environment(t *testing.T) {
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
			withApp010Fs(t, "app010_app.yaml", func(fs afero.Fs) {
				app, err := NewApp010(fs, "/")
				require.NoError(t, err)

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

func TestApp010_AddEnvironment(t *testing.T) {
	withApp010Fs(t, "app010_app.yaml", func(fs afero.Fs) {
		app, err := NewApp010(fs, "/")
		require.NoError(t, err)

		envs, err := app.Environments()
		require.NoError(t, err)

		envLen := len(envs)

		newEnv := &EnvironmentSpec{
			Destination: &EnvironmentDestinationSpec{
				Namespace: "some-namespace",
				Server:    "http://example.com",
			},
			Path: "us-west/qa",
		}

		k8sSpecFlag := "version:v1.8.7"
		err = app.AddEnvironment("us-west/qa", k8sSpecFlag, newEnv)
		require.NoError(t, err)

		envs, err = app.Environments()
		require.NoError(t, err)
		require.Len(t, envs, envLen+1)

		env, err := app.Environment("us-west/qa")
		require.NoError(t, err)
		require.Equal(t, "v1.8.7", env.KubernetesVersion)
	})
}

func TestApp010_RemoveEnvironment(t *testing.T) {
	withApp010Fs(t, "app010_app.yaml", func(fs afero.Fs) {
		app, err := NewApp010(fs, "/")
		require.NoError(t, err)

		_, err = app.Environment("default")
		require.NoError(t, err)

		err = app.RemoveEnvironment("default")
		require.NoError(t, err)

		app, err = NewApp010(fs, "/")
		require.NoError(t, err)

		_, err = app.Environment("default")
		require.Error(t, err)
	})
}

func withApp010Fs(t *testing.T, appName string, fn func(fs afero.Fs)) {
	ogLibUpdater := LibUpdater
	LibUpdater = func(fs afero.Fs, k8sSpecFlag string, libPath string, useVersionPath bool) (string, error) {
		return "v1.8.7", nil
	}

	defer func() {
		LibUpdater = ogLibUpdater
	}()

	fs := afero.NewMemMapFs()
	stageFile(t, fs, appName, "/app.yaml")

	fn(fs)
}

func stageFile(t *testing.T, fs afero.Fs, src, dest string) {
	in := filepath.Join("testdata", src)

	b, err := ioutil.ReadFile(in)
	require.NoError(t, err)

	dir := filepath.Dir(dest)
	err = fs.MkdirAll(dir, 0755)
	require.NoError(t, err)

	err = afero.WriteFile(fs, dest, b, 0644)
	require.NoError(t, err)
}
