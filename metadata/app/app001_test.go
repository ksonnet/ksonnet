package app

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApp001_RenameEnvironment(t *testing.T) {
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
				"/environments/renamed/.metadata",
				"/environments/renamed/spec.json",
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
				"/environments/default/nested/.metadata",
				"/environments/default/nested/spec.json",
			},
			shouldNotExist: []string{
				"/environments/default/.metadata",
			},
		},
	}

	for _, tc := range cases {
		withApp001Fs(t, "app001_app.yaml", func(fs afero.Fs) {
			t.Run(tc.name, func(t *testing.T) {
				app, err := NewApp001(fs, "/")
				require.NoError(t, err)

				err = app.RenameEnvironment(tc.from, tc.to)
				require.NoError(t, err)

				for _, p := range tc.shouldExist {
					checkExist(t, fs, p)
				}

				for _, p := range tc.shouldNotExist {
					checkNotExist(t, fs, p)
				}

				app.load()
				require.NoError(t, err)

				_, err = app.Environment(tc.from)
				assert.Error(t, err)

				_, err = app.Environment(tc.to)
				assert.NoError(t, err)
			})
		})
	}
}

func TestApp001_Environments(t *testing.T) {
	withApp001Fs(t, "app001_app.yaml", func(fs afero.Fs) {
		app, err := NewApp001(fs, "/")
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
				KubernetesVersion: "v1.7.0",
				Path:              "us-east/test",
			},
			"us-west/test": &EnvironmentSpec{
				Destination: &EnvironmentDestinationSpec{
					Namespace: "some-namespace",
					Server:    "http://example.com",
				},
				KubernetesVersion: "v1.7.0",
				Path:              "us-west/test",
			},
			"us-west/prod": &EnvironmentSpec{
				Destination: &EnvironmentDestinationSpec{
					Namespace: "some-namespace",
					Server:    "http://example.com",
				},
				KubernetesVersion: "v1.7.0",
				Path:              "us-west/prod",
			},
		}
		envs, err := app.Environments()
		require.NoError(t, err)

		require.Equal(t, expected, envs)
	})
}

func TestApp001_Environment(t *testing.T) {
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
			withApp001Fs(t, "app001_app.yaml", func(fs afero.Fs) {
				app, err := NewApp001(fs, "/")
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

func TestApp001_AddEnvironment(t *testing.T) {
	withApp001Fs(t, "app001_app.yaml", func(fs afero.Fs) {
		app, err := NewApp001(fs, "/")
		require.NoError(t, err)

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

		_, err = app.Environment("us-west/qa")
		require.NoError(t, err)
	})
}

func TestApp001_Upgrade_dryrun(t *testing.T) {
	withApp001Fs(t, "app001_app.yaml", func(fs afero.Fs) {
		app, err := NewApp001(fs, "/")
		require.NoError(t, err)

		var buf bytes.Buffer
		app.out = &buf

		err = app.Upgrade(true)
		require.NoError(t, err)
	})
}

func TestApp001_Upgrade(t *testing.T) {
	withApp001Fs(t, "app001_app.yaml", func(fs afero.Fs) {
		app, err := NewApp001(fs, "/")
		require.NoError(t, err)

		var buf bytes.Buffer
		app.out = &buf

		err = app.Upgrade(false)
		require.NoError(t, err)

		root := filepath.Join(app.root, EnvironmentDirName)
		var foundSpec bool
		err = afero.Walk(fs, root, func(path string, fi os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if fi.IsDir() {
				return nil
			}

			if fi.Name() == "spec.json" {
				foundSpec = true
			}
			return nil
		})

		require.NoError(t, err)
		require.False(t, foundSpec)
	})
}

func withApp001Fs(t *testing.T, appName string, fn func(fs afero.Fs)) {
	ogLibUpdater := LibUpdater
	LibUpdater = func(fs afero.Fs, k8sSpecFlag string, libPath string, useVersionPath bool) (string, error) {
		path := filepath.Join(libPath, "swagger.json")
		stageFile(t, fs, "swagger.json", path)
		return "v1.8.7", nil
	}

	defer func() {
		LibUpdater = ogLibUpdater
	}()

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
		err := fs.MkdirAll(path, DefaultFolderPermissions)
		require.NoError(t, err)

		specPath := filepath.Join(path, "spec.json")
		stageFile(t, fs, "spec.json", specPath)

		swaggerPath := filepath.Join(path, ".metadata", "swagger.json")
		stageFile(t, fs, "swagger.json", swaggerPath)
	}

	stageFile(t, fs, appName, "/app.yaml")

	fn(fs)
}
