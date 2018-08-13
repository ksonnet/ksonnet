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

package actions

import (
	"bytes"
	"io"
	"testing"

	"github.com/ksonnet/ksonnet/pkg/app"
	"github.com/ksonnet/ksonnet/pkg/prototype"
	"github.com/pkg/errors"

	amocks "github.com/ksonnet/ksonnet/pkg/app/mocks"
	pkgmocks "github.com/ksonnet/ksonnet/pkg/pkg/mocks"
	"github.com/ksonnet/ksonnet/pkg/registry"
	regmocks "github.com/ksonnet/ksonnet/pkg/registry/mocks"
	"github.com/ksonnet/ksonnet/pkg/util/test"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

func TestPkgDescribe(t *testing.T) {

	cases := []struct {
		name        string
		output      string
		pkgManager  func() registry.PackageManager
		isErr       bool
		templateSrc string
		out         io.Writer
	}{
		{
			name:   "with no prototypes",
			output: "pkg/describe/output.txt",
			pkgManager: func() registry.PackageManager {
				p := &pkgmocks.Package{}
				p.On("Description").Return("description")
				p.On("IsInstalled").Return(false, nil)

				pkgManager := &regmocks.PackageManager{}
				pkgManager.On("Find", "apache").Return(p, nil)

				return pkgManager
			},
		},
		{
			name:   "with prototypes",
			output: "pkg/describe/with-prototypes.txt",
			pkgManager: func() registry.PackageManager {
				prototypes := prototype.Prototypes{
					{
						Name: "proto1",
						Template: prototype.SnippetSchema{
							ShortDescription: "short description",
						},
					},
				}

				p := &pkgmocks.Package{}
				p.On("Description").Return("description")
				p.On("IsInstalled").Return(true, nil)
				p.On("Prototypes").Return(prototypes, nil)

				pkgManager := &regmocks.PackageManager{}
				pkgManager.On("Find", "apache").Return(p, nil)

				return pkgManager
			},
		},
		{
			name:  "package manager find error",
			isErr: true,
			pkgManager: func() registry.PackageManager {
				pkgManager := &regmocks.PackageManager{}
				pkgManager.On("Find", "apache").Return(nil, errors.New("failed"))

				return pkgManager
			},
		},
		{
			name:  "check installed returns error",
			isErr: true,
			pkgManager: func() registry.PackageManager {
				p := &pkgmocks.Package{}
				p.On("Description").Return("description")
				p.On("IsInstalled").Return(false, errors.New("failed"))

				pkgManager := &regmocks.PackageManager{}
				pkgManager.On("Find", "apache").Return(p, nil)

				return pkgManager
			},
		},
		{
			name:  "gather prototypes returns error",
			isErr: true,
			pkgManager: func() registry.PackageManager {
				p := &pkgmocks.Package{}
				p.On("Prototypes").Return(nil, errors.New("failed"))
				p.On("Description").Return("description")
				p.On("IsInstalled").Return(true, nil)

				pkgManager := &regmocks.PackageManager{}
				pkgManager.On("Find", "apache").Return(p, nil)

				return pkgManager
			},
		},
		{
			name:  "template is invalid",
			isErr: true,
			pkgManager: func() registry.PackageManager {
				p := &pkgmocks.Package{}
				p.On("Description").Return("description")
				p.On("IsInstalled").Return(false, nil)

				pkgManager := &regmocks.PackageManager{}
				pkgManager.On("Find", "apache").Return(p, nil)

				return pkgManager
			},
			templateSrc: "{{-abc}}",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Fatal(r)
				}
			}()

			test.WithApp(t, "/app", func(a *amocks.App, fs afero.Fs) {
				libraries := app.LibraryConfigs{
					"apache": &app.LibraryConfig{},
				}

				a.On("Libraries").Return(libraries, nil)

				in := map[string]interface{}{
					OptionApp:           a,
					OptionPackageName:   "apache",
					OptionTLSSkipVerify: false,
				}

				pd, err := NewPkgDescribe(in)
				require.NoError(t, err)

				if tc.templateSrc != "" {
					pd.templateSrc = tc.templateSrc
				}

				pd.packageManager = tc.pkgManager()

				var buf bytes.Buffer
				pd.out = &buf

				err = pd.Run()
				if tc.isErr {
					require.Error(t, err)
					return
				}
				require.NoError(t, err)

				assertOutput(t, tc.output, buf.String())
			})
		})
	}
}

func TestPkgDescribe_requires_app(t *testing.T) {
	in := make(map[string]interface{})
	_, err := NewPkgDescribe(in)
	require.Error(t, err)
}
