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

package registry

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ksonnet/ksonnet/pkg/app"
	"github.com/ksonnet/ksonnet/pkg/app/mocks"
	"github.com/ksonnet/ksonnet/pkg/parts"
	"github.com/ksonnet/ksonnet/pkg/util/test"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func withRFS(t *testing.T, relPath bool, fn func(*Fs, *mocks.App, afero.Fs)) {
	uri := "/work/local"
	if relPath {
		uri = "../work/local"
	}

	fs := afero.NewMemMapFs()
	appMock := &mocks.App{}
	appMock.On("Fs").Return(fs)
	appMock.On("Root").Return("/app")
	appMock.On("LibPath", mock.AnythingOfType("string")).Return(filepath.Join("/app", "lib", "v1.8.7"), nil)

	spec := &app.RegistryConfig{
		Name:     "local",
		Protocol: string(ProtocolFilesystem),
		URI:      uri,
	}

	partRoot := filepath.Join("testdata", "part", "incubator")
	err := filepath.Walk(partRoot, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		newPath := filepath.Join("/work", "local", strings.TrimPrefix(path, partRoot))
		if fi.IsDir() {
			return fs.MkdirAll(newPath, 0750)
		}

		data, err := ioutil.ReadFile(path)
		require.NoError(t, err)

		return afero.WriteFile(fs, newPath, data, 0644)
	})
	require.NoError(t, err)

	data, err := ioutil.ReadFile(filepath.Join("testdata", "fs-registry.yaml"))
	require.NoError(t, err)

	err = afero.WriteFile(fs, "/work/local/registry.yaml", data, 0644)
	require.NoError(t, err)

	rfs, err := NewFs(appMock, spec)
	require.NoError(t, err)

	fn(rfs, appMock, fs)
}

func TestFs_ValidateURI_invalid_path(t *testing.T) {
	test.WithApp(t, "/app", func(a *mocks.App, fs afero.Fs) {
		uri := "/invalid"

		spec := &app.RegistryConfig{
			Name:     "local",
			Protocol: string(ProtocolFilesystem),
			URI:      uri,
		}
		r, err := NewFs(a, spec)
		require.NoError(t, err)

		ok, err := r.ValidateURI(uri)
		require.Error(t, err)
		assert.Equal(t, false, ok)
	})
}

func TestFs_Name(t *testing.T) {
	withRFS(t, false, func(rfs *Fs, appMock *mocks.App, fs afero.Fs) {
		assert.Equal(t, "local", rfs.Name())
	})
}

func TestFs_Protocol(t *testing.T) {
	withRFS(t, false, func(rfs *Fs, appMock *mocks.App, fs afero.Fs) {
		assert.Equal(t, ProtocolFilesystem, rfs.Protocol())
	})
}

func TestFs_URI(t *testing.T) {
	withRFS(t, false, func(rfs *Fs, appMock *mocks.App, fs afero.Fs) {
		assert.Equal(t, "/work/local", rfs.URI())
	})
}

func TestFs_RegistrySpecDir(t *testing.T) {
	withRFS(t, false, func(rfs *Fs, appMock *mocks.App, fs afero.Fs) {
		assert.Equal(t, "/work/local", rfs.RegistrySpecDir())
	})
}

func TestFs_RegistrySpecFilePath(t *testing.T) {
	withRFS(t, false, func(rfs *Fs, appMock *mocks.App, fs afero.Fs) {
		assert.Equal(t, "/work/local/registry.yaml", rfs.RegistrySpecFilePath())
	})
}

func TestFs_FetchRegistrySpec(t *testing.T) {
	cases := []struct {
		name    string
		relPath bool
	}{
		{
			name:    "relative path",
			relPath: true,
		},
		{
			name: "absolute path",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			withRFS(t, tc.relPath, func(rfs *Fs, appMock *mocks.App, fs afero.Fs) {
				spec, err := rfs.FetchRegistrySpec()
				require.NoError(t, err)

				expected := &Spec{
					APIVersion: "0.1.0",
					Kind:       "ksonnet.io/registry",
					Libraries: LibraryConfigs{
						"apache": &LibaryConfig{
							Path: "apache",
						},
					},
				}

				assert.Equal(t, expected, spec)
			})
		})
	}

}

func TestFs_MakeRegistryConfig(t *testing.T) {
	withRFS(t, false, func(rfs *Fs, appMock *mocks.App, fs afero.Fs) {
		expected := &app.RegistryConfig{
			Name:     "local",
			Protocol: string(ProtocolFilesystem),
			URI:      "/work/local",
		}
		assert.Equal(t, expected, rfs.MakeRegistryConfig())

	})
}

func TestFs_ResolveLibrarySpec(t *testing.T) {
	withRFS(t, false, func(rfs *Fs, appMock *mocks.App, fs afero.Fs) {
		spec, err := rfs.ResolveLibrarySpec("apache", "")
		require.NoError(t, err)

		expected := &parts.Spec{
			APIVersion:  "0.0.1",
			Kind:        "ksonnet.io/parts",
			Name:        "apache",
			Description: "part description",
			Author:      "author",
			Contributors: parts.ContributorSpecs{
				&parts.ContributorSpec{Name: "author 1", Email: "email@example.com"},
				&parts.ContributorSpec{Name: "author 2", Email: "email@example.com"},
			},
			Repository: parts.RepositorySpec{
				Type: "git",
				URL:  "https://github.com/ksonnet/mixins",
			},
			Bugs: &parts.BugSpec{
				URL: "https://github.com/ksonnet/mixins/issues",
			},
			Keywords: []string{"apache", "server", "http"},
			QuickStart: &parts.QuickStartSpec{
				Prototype:     "io.ksonnet.pkg.apache-simple",
				ComponentName: "apache",
				Flags: map[string]string{
					"name":      "apache",
					"namespace": "default",
				},
				Comment: "Run a simple Apache server",
			},
			License: "Apache 2.0",
		}

		assert.Equal(t, expected, spec)

	})
}

func TestFs_ResolveLibrary(t *testing.T) {
	withRFS(t, false, func(rfs *Fs, appMock *mocks.App, fs afero.Fs) {
		var files []string
		onFile := func(relPath string, contents []byte) error {
			files = append(files, relPath)
			return nil
		}

		var directories []string
		onDir := func(relPath string) error {
			directories = append(directories, relPath)
			return nil
		}

		spec, libRefSpec, err := rfs.ResolveLibrary("apache", "alias", "54321", onFile, onDir)
		require.NoError(t, err)

		expectedSpec := &parts.Spec{
			APIVersion:  "0.0.1",
			Kind:        "ksonnet.io/parts",
			Name:        "apache",
			Description: "part description",
			Author:      "author",
			Contributors: parts.ContributorSpecs{
				&parts.ContributorSpec{Name: "author 1", Email: "email@example.com"},
				&parts.ContributorSpec{Name: "author 2", Email: "email@example.com"},
			},
			Repository: parts.RepositorySpec{
				Type: "git",
				URL:  "https://github.com/ksonnet/mixins",
			},
			Bugs: &parts.BugSpec{
				URL: "https://github.com/ksonnet/mixins/issues",
			},
			Keywords: []string{"apache", "server", "http"},
			QuickStart: &parts.QuickStartSpec{
				Prototype:     "io.ksonnet.pkg.apache-simple",
				ComponentName: "apache",
				Flags: map[string]string{
					"name":      "apache",
					"namespace": "default",
				},
				Comment: "Run a simple Apache server",
			},
			License: "Apache 2.0",
		}
		assert.Equal(t, expectedSpec, spec)

		expectedLibRefSpec := &app.LibraryConfig{
			Name:     "alias",
			Registry: "local",
		}
		assert.Equal(t, expectedLibRefSpec, libRefSpec)

		expectedFiles := []string{
			"apache/README.md",
			"apache/apache.libsonnet",
			"apache/examples/apache.jsonnet",
			"apache/examples/generated.yaml",
			"apache/parts.yaml",
			"apache/prototypes/apache-simple.jsonnet",
		}
		assert.Equal(t, expectedFiles, files)

		expectedDirs := []string{
			"apache",
			"apache/examples",
			"apache/prototypes",
		}
		assert.Equal(t, expectedDirs, directories)

	})
}
