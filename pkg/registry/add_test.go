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
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"

	"github.com/ksonnet/ksonnet/pkg/app"
	amocks "github.com/ksonnet/ksonnet/pkg/app/mocks"
	"github.com/ksonnet/ksonnet/pkg/helm"
	helmmocks "github.com/ksonnet/ksonnet/pkg/helm/mocks"
	ghutil "github.com/ksonnet/ksonnet/pkg/util/github"
	"github.com/ksonnet/ksonnet/pkg/util/github/mocks"
	"github.com/ksonnet/ksonnet/pkg/util/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func withApp(t *testing.T, fn func(*amocks.App, afero.Fs)) {
	ogGHF := githubFactory
	defer func(fn ghFactoryFn) {
		githubFactory = fn
	}(ogGHF)

	ogHF := helmFactory
	defer func(fn helmFactoryFn) {
		helmFactory = fn
	}(ogHF)

	test.WithApp(t, "/app", func(a *amocks.App, fs afero.Fs) {
		fn(a, fs)
	})

}

func TestAdd_GitHub(t *testing.T) {
	withApp(t, func(appMock *amocks.App, fs afero.Fs) {
		expectedSpec := &app.RegistryConfig{
			Name:     "new",
			Protocol: string(ProtocolGitHub),
			URI:      "github.com/foo/bar",
		}

		appMock.On("AddRegistry", expectedSpec, true).Return(nil)

		ghMock := &mocks.GitHub{}
		ghMock.On("ValidateURL", "github.com/foo/bar").Return(nil)
		ghMock.On("CommitSHA1", mock.Anything, mock.Anything, "master").Return("40285d8a14f1ac5787e405e1023cf0c07f6aa28c", nil)

		registryContent := buildContent(t, registryYAMLFile)
		ghMock.On(
			"Contents",
			mock.Anything,
			ghutil.Repo{Org: "foo", Repo: "bar"},
			registryYAMLFile,
			"40285d8a14f1ac5787e405e1023cf0c07f6aa28c").
			Return(registryContent, nil, nil)

		ghOpt := GitHubClient(ghMock)
		githubFactory = func(a app.App, registryRef *app.RegistryConfig, opts ...GitHubOpt) (*GitHub, error) {
			return NewGitHub(a, registryRef, ghOpt)
		}

		spec, err := Add(appMock, ProtocolGitHub, "new", "github.com/foo/bar", true, nil)
		require.NoError(t, err)

		require.Equal(t, registrySpec, spec)

	})
}

func TestAdd_Helm(t *testing.T) {
	withApp(t, func(appMock *amocks.App, fs afero.Fs) {
		expectedRegistryConfig := &app.RegistryConfig{
			Name:     "new",
			Protocol: string(ProtocolHelm),
			URI:      "http://example.com",
		}

		appMock.On("AddRegistry", expectedRegistryConfig, false).Return(nil)

		f, err := os.Open(filepath.Join("testdata", "helm-index.yaml"))
		require.NoError(t, err)

		getterMock := &helmmocks.Getter{}
		resp := &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(f),
		}
		getterMock.On("Get", "http://example.com/index.yaml").Return(resp, nil)

		httpClient, err := helm.NewHTTPClient("http://example.com", getterMock)
		require.NoError(t, err)

		helmFactory = func(a app.App, registryConfig *app.RegistryConfig, _ *helm.HTTPClient) (*Helm, error) {
			return NewHelm(a, registryConfig, httpClient, nil)
		}

		spec, err := Add(appMock, ProtocolHelm, "new", "http://example.com", false, nil)
		require.NoError(t, err)

		expected := &Spec{
			APIVersion: DefaultAPIVersion,
			Kind:       DefaultKind,
			Libraries: LibraryConfigs{
				"chart": &LibraryConfig{
					Version: "2.0.0",
					Path:    "chart",
				},
			},
		}

		require.Equal(t, expected, spec)
	})
}

func TestAdd_fs(t *testing.T) {
	test.WithApp(t, "/app", func(a *amocks.App, fs afero.Fs) {
		_, err := Add(a, ProtocolFilesystem, "/invalid", "", false, nil)
		require.Error(t, err)
	})
}

func TestAdd_invalid(t *testing.T) {
	test.WithApp(t, "/app", func(a *amocks.App, fs afero.Fs) {
		_, err := Add(a, Protocol("invalid"), "", "", false, nil)
		require.Error(t, err)
	})
}

func Test_load(t *testing.T) {
	withApp(t, func(appMock *amocks.App, fs afero.Fs) {

		test.StageFile(t, fs, "registry.yaml", "/app/registry.yaml")
		test.StageFile(t, fs, "invalid-registry.yaml", "/app/invalid-registry.yaml")

		cases := []struct {
			name     string
			path     string
			isErr    bool
			expected *Spec
			exists   bool
		}{
			{
				name:     "file exists",
				path:     "/app/registry.yaml",
				expected: registrySpec,
				exists:   true,
			},
			{
				name:   "file is not valid",
				path:   "/app/invalid-registry.yaml",
				isErr:  true,
				exists: false,
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				s, exists, err := load(appMock, tc.path)

				if tc.isErr {
					require.Error(t, err)
					return
				}

				require.NoError(t, err)
				assert.Equal(t, tc.expected, s)
				assert.Equal(t, tc.exists, exists)
			})
		}

	})
}
