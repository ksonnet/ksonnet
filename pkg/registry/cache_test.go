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
	"testing"

	"github.com/ksonnet/ksonnet/metadata/app"
	amocks "github.com/ksonnet/ksonnet/metadata/app/mocks"
	"github.com/ksonnet/ksonnet/pkg/pkg"
	ghutil "github.com/ksonnet/ksonnet/pkg/util/github"
	"github.com/ksonnet/ksonnet/pkg/util/github/mocks"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_CacheDependency(t *testing.T) {
	withApp(t, func(a *amocks.App, fs afero.Fs) {
		libraries := app.LibraryRefSpecs{}
		a.On("Libraries").Return(libraries, nil)

		registries := app.RegistryRefSpecs{
			"incubator": &app.RegistryRefSpec{
				Name:     "incubator",
				Protocol: ProtocolGitHub,
				URI:      "github.com/foo/bar/tree/master/incubator",
				GitVersion: &app.GitVersionSpec{
					CommitSHA: "54321",
					RefSpec:   "master",
				},
			},
		}
		a.On("Registries").Return(registries, nil)

		ghMock := &mocks.GitHub{}
		ghMock.On("CommitSHA1", mock.Anything, mock.Anything, "master").Return("54321", nil)

		repo := ghutil.Repo{Org: "foo", Repo: "bar"}
		mockPartFs(t, repo, ghMock, "incubator/apache", "54321")

		registryContent := buildContent(t, registryYAMLFile)
		ghMock.On(
			"Contents",
			mock.Anything,
			registryYAMLFile,
			"40285d8a14f1ac5787e405e1023cf0c07f6aa28c").
			Return(registryContent, nil, nil)

		ghOpt := GitHubClient(ghMock)
		githubFactory = func(registryRef *app.RegistryRefSpec) (*GitHub, error) {
			return NewGitHub(registryRef, ghOpt)
		}

		library := &app.LibraryRefSpec{
			Name:     "apache",
			Registry: "incubator",
			GitVersion: &app.GitVersionSpec{
				CommitSHA: "54321",
				RefSpec:   "master",
			},
		}
		a.On("UpdateLib", "apache", library).Return(nil)

		d := pkg.Descriptor{Registry: "incubator", Part: "apache"}

		err := CacheDependency(a, d, "")
		require.NoError(t, err)
	})
}
