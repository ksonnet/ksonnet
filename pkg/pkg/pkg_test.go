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

package pkg

import (
	"errors"
	"testing"

	"github.com/ksonnet/ksonnet/pkg/app"
	amocks "github.com/ksonnet/ksonnet/pkg/app/mocks"
	"github.com/ksonnet/ksonnet/pkg/util/test"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

func Test_DefaultInstallChecker_isInstalled(t *testing.T) {
	cases := []struct {
		name           string
		setupLibraries func(*amocks.App)
		isInstalled    bool
		isErr          bool
	}{
		{
			name: "is installed",
			setupLibraries: func(a *amocks.App) {
				libraries := app.LibraryRefSpecs{
					"redis": &app.LibraryRefSpec{},
				}

				a.On("Libraries").Return(libraries, nil)
			},
			isInstalled: true,
		},
		{
			name: "not installed",
			setupLibraries: func(a *amocks.App) {
				a.On("Libraries").Return(nil, nil)
			},
		},
		{
			name: "libraries error",
			setupLibraries: func(a *amocks.App) {
				a.On("Libraries").Return(nil, errors.New("failed"))
			},
			isErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			test.WithApp(t, "/app", func(a *amocks.App, fs afero.Fs) {
				tc.setupLibraries(a)

				ic := DefaultInstallChecker{App: a}

				i, err := ic.IsInstalled("redis")
				if tc.isErr {
					require.Error(t, err)
					return
				}

				require.NoError(t, err)
				require.Equal(t, tc.isInstalled, i)
			})
		})
	}
}

type fakeInstallChecker struct {
	isInstalled    bool
	isInstalledErr error
}

func (ic *fakeInstallChecker) IsInstalled(name string) (bool, error) {
	return ic.isInstalled, ic.isInstalledErr
}
