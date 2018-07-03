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
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ksonnet/ksonnet/pkg/app"
	"github.com/ksonnet/ksonnet/pkg/app/mocks"
	"github.com/ksonnet/ksonnet/pkg/helm"
	"github.com/ksonnet/ksonnet/pkg/parts"
	"github.com/ksonnet/ksonnet/pkg/util/archive"
	"github.com/ksonnet/ksonnet/pkg/util/test"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHelm_requires_repository_client(t *testing.T) {
	test.WithApp(t, "/app", func(a *mocks.App, fs afero.Fs) {
		spec := &app.RegistryConfig{
			URI:  "http://example.com",
			Name: "name",
		}
		_, err := NewHelm(a, spec, nil, nil)
		require.Error(t, err)
	})
}

func TestHelm_requires_name(t *testing.T) {
	test.WithApp(t, "/app", func(a *mocks.App, fs afero.Fs) {
		spec := &app.RegistryConfig{
			URI:  "http://example.com",
			Name: "name",
		}
		h, err := NewHelm(a, spec, &fakeHelmRepositoryClient{}, nil)
		require.NoError(t, err)

		require.Equal(t, "name", h.Name())
	})
}

func TestHelm_requires_uri(t *testing.T) {
	test.WithApp(t, "/app", func(a *mocks.App, fs afero.Fs) {
		spec := &app.RegistryConfig{
			URI:  "http://example.com",
			Name: "name",
		}
		h, err := NewHelm(a, spec, &fakeHelmRepositoryClient{}, nil)
		require.NoError(t, err)

		require.Equal(t, "http://example.com", h.URI())
	})
}

func TestHelm_requires_protocol(t *testing.T) {
	test.WithApp(t, "/app", func(a *mocks.App, fs afero.Fs) {
		spec := &app.RegistryConfig{
			URI:  "http://example.com",
			Name: "name",
		}
		h, err := NewHelm(a, spec, &fakeHelmRepositoryClient{}, nil)
		require.NoError(t, err)

		require.Equal(t, ProtocolHelm, h.Protocol())
	})
}

func TestHelm_RegistrySpecDir(t *testing.T) {
	test.WithApp(t, "/app", func(a *mocks.App, fs afero.Fs) {
		spec := &app.RegistryConfig{
			URI:  "http://example.com",
			Name: "name",
		}
		h, err := NewHelm(a, spec, &fakeHelmRepositoryClient{}, nil)
		require.NoError(t, err)

		require.Equal(t, "name", h.RegistrySpecDir())
	})
}

func TestHelm_RegistrySpecFilePath(t *testing.T) {
	test.WithApp(t, "/app", func(a *mocks.App, fs afero.Fs) {
		spec := &app.RegistryConfig{
			URI:  "http://example.com",
			Name: "name",
		}
		h, err := NewHelm(a, spec, &fakeHelmRepositoryClient{}, nil)
		require.NoError(t, err)

		require.Equal(t, "", h.RegistrySpecFilePath())
	})
}
func TestHelm_FetchRegistrySpec(t *testing.T) {
	test.WithApp(t, "/app", func(a *mocks.App, fs afero.Fs) {
		spec := &app.RegistryConfig{
			URI:  "http://example.com",
			Name: "name",
		}

		rc := &fakeHelmRepositoryClient{
			entries: &helm.Repository{
				Charts: map[string][]helm.RepositoryChart{
					"app-a": []helm.RepositoryChart{
						{
							Name:    "app-a",
							Version: "0.1.0",
						},
					},
					"app-b": []helm.RepositoryChart{
						{
							Name:    "app-b",
							Version: "0.1.0",
						},
						{
							Name:    "app-b",
							Version: "0.2.0",
						},
					},
				},
			},
		}

		h, err := NewHelm(a, spec, rc, nil)
		require.NoError(t, err)

		got, err := h.FetchRegistrySpec()
		require.NoError(t, err)

		expected := &Spec{
			APIVersion: DefaultAPIVersion,
			Kind:       DefaultKind,
			Libraries: LibraryConfigs{
				"app-a": &LibaryConfig{
					Path:    "app-a",
					Version: "0.1.0",
				},
				"app-b": &LibaryConfig{
					Path:    "app-b",
					Version: "0.2.0",
				},
			},
		}

		require.Equal(t, expected, got)
	})
}

func TestHelm_MakeRegistryConfig(t *testing.T) {
	test.WithApp(t, "/app", func(a *mocks.App, fs afero.Fs) {
		spec := &app.RegistryConfig{
			URI:  "http://example.com",
			Name: "name",
		}
		h, err := NewHelm(a, spec, &fakeHelmRepositoryClient{}, nil)
		require.NoError(t, err)

		got := h.MakeRegistryConfig()

		require.Equal(t, spec, got)
	})
}

func TestHelm_ResolveLibrarySpec(t *testing.T) {
	test.WithApp(t, "/app", func(a *mocks.App, fs afero.Fs) {
		spec := &app.RegistryConfig{
			URI:  "http://example.com",
			Name: "name",
		}

		rc := &fakeHelmRepositoryClient{
			chart: &helm.RepositoryChart{
				Name:        "app-a",
				Version:     "0.1.0",
				Description: "description",
			},
		}

		h, err := NewHelm(a, spec, rc, nil)
		require.NoError(t, err)

		part, err := h.ResolveLibrarySpec("app-a", "")
		require.NoError(t, err)

		expected := &parts.Spec{
			APIVersion:  parts.DefaultAPIVersion,
			Kind:        parts.DefaultKind,
			Name:        "app-a",
			Version:     "0.1.0",
			Description: "description",
		}

		require.Equal(t, expected, part)
	})
}

func TestHelm_ResolveLibrarySpec_chart_error(t *testing.T) {
	test.WithApp(t, "/app", func(a *mocks.App, fs afero.Fs) {
		spec := &app.RegistryConfig{
			URI:  "http://example.com",
			Name: "name",
		}

		rc := &fakeHelmRepositoryClient{
			chartErr: errors.New("error"),
		}

		h, err := NewHelm(a, spec, rc, nil)
		require.NoError(t, err)

		_, err = h.ResolveLibrarySpec("app-a", "")
		require.Error(t, err)
	})
}

func TestHelm_ResolveLibrary(t *testing.T) {
	test.WithApp(t, "/app", func(a *mocks.App, fs afero.Fs) {
		spec := &app.RegistryConfig{
			URI:  "http://example.com",
			Name: "name",
		}

		rc := &fakeHelmRepositoryClient{
			chart: &helm.RepositoryChart{
				Name:        "app-a",
				Version:     "0.1.0",
				Description: "description",
				URLs:        []string{"http://example.com/archive"},
			},

			fetchReader: ioutil.NopCloser(strings.NewReader("contents")),
		}

		unarchiver := &fakeUnarchiver{}
		h, err := NewHelm(a, spec, rc, unarchiver)
		require.NoError(t, err)

		var foundFile bool
		fileHandler := func(relPath string, contents []byte) error {
			if relPath == "app-a/helm/0.1.0/part/README.md" {
				assert.Equal(t, "hello world", string(contents))
				foundFile = true
			}
			return nil
		}

		dirHandler := func(relPath string) error {
			return nil
		}

		part, lib, err := h.ResolveLibrary("app-a", "", "", fileHandler, dirHandler)
		require.NoError(t, err)

		expectedPart := &parts.Spec{
			APIVersion:  parts.DefaultAPIVersion,
			Kind:        parts.DefaultKind,
			Name:        "app-a",
			Version:     "0.1.0",
			Description: "description",
		}

		assert.Equal(t, expectedPart, part)

		expectedLib := &app.LibraryConfig{
			Registry: "name",
		}

		assert.Equal(t, expectedLib, lib)
		assert.True(t, foundFile)
	})
}

func TestHelm_CacheRoot(t *testing.T) {
	test.WithApp(t, "/app", func(a *mocks.App, fs afero.Fs) {
		spec := &app.RegistryConfig{
			URI:  "http://example.com",
			Name: "name",
		}

		rc := &fakeHelmRepositoryClient{
			chartErr: errors.New("error"),
		}

		h, err := NewHelm(a, spec, rc, nil)
		require.NoError(t, err)

		got, err := h.CacheRoot("name", "path")
		require.NoError(t, err)

		require.Equal(t, "name/path", filepath.ToSlash(got))
	})
}

func TestHelm_ValidateURI_invalid(t *testing.T) {
	test.WithApp(t, "/app", func(a *mocks.App, fs afero.Fs) {
		uri := ""

		spec := &app.RegistryConfig{
			Name:     "local",
			Protocol: string(ProtocolHelm),
			URI:      uri,
		}

		rc := &fakeHelmRepositoryClient{
			chartErr: errors.New("error"),
		}

		h, err := NewHelm(a, spec, rc, nil)
		require.NoError(t, err)

		ok, err := h.ValidateURI(uri)
		require.Error(t, err)
		assert.Equal(t, false, ok)
	})
}

type fakeHelmRepositoryClient struct {
	entries    *helm.Repository
	entriesErr error

	chart    *helm.RepositoryChart
	chartErr error

	fetchReader io.ReadCloser
	fetchErr    error
}

func (hrc *fakeHelmRepositoryClient) Repository() (*helm.Repository, error) {
	return hrc.entries, hrc.entriesErr
}

func (hrc *fakeHelmRepositoryClient) Chart(string, string) (*helm.RepositoryChart, error) {
	return hrc.chart, hrc.chartErr
}

func (hrc *fakeHelmRepositoryClient) Fetch(s string) (io.ReadCloser, error) {
	fmt.Println("fetching", s)
	return hrc.fetchReader, hrc.fetchErr
}

type fakeUnarchiver struct{}

func (u *fakeUnarchiver) Unarchive(_ io.Reader, h archive.FileHandler) error {
	f := &archive.File{
		Reader: strings.NewReader("hello world"),
		Name:   "part/README.md",
	}

	return h(f)
}
