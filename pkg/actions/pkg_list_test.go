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
	"fmt"
	"testing"

	"github.com/ksonnet/ksonnet/pkg/app"
	amocks "github.com/ksonnet/ksonnet/pkg/app/mocks"
	"github.com/ksonnet/ksonnet/pkg/pkg"
	pmocks "github.com/ksonnet/ksonnet/pkg/pkg/mocks"
	"github.com/ksonnet/ksonnet/pkg/prototype"
	rmocks "github.com/ksonnet/ksonnet/pkg/registry/mocks"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type myPkg struct {
	name        string
	registry    string
	version     string
	isInstalled bool
}

func (p myPkg) Name() string {
	return p.name
}
func (p myPkg) Version() string {
	return p.version
}
func (p myPkg) RegistryName() string {
	return p.registry
}
func (p myPkg) Description() string {
	return ""
}
func (p myPkg) Path() string {
	return p.name
}
func (p myPkg) String() string {
	return fmt.Sprintf("%s/%s@%s", p.registry, p.name, p.version)
}
func (p myPkg) Prototypes() (prototype.Prototypes, error) {
	return nil, errors.New("not implemented")
}
func (p myPkg) IsInstalled() (bool, error) {
	return p.isInstalled, nil
}

var _ = pkg.Package(myPkg{})

func TestPkgList(t *testing.T) {
	withApp(t, func(appMock *amocks.App) {
		pmMock := rmocks.PackageManager{}
		pmMock.On("Packages").Return(
			[]pkg.Package{
				myPkg{
					name:        "lib1",
					version:     "0.0.1",
					registry:    "incubator",
					isInstalled: true,
				},
				myPkg{
					name:        "lib1",
					version:     "0.0.2",
					registry:    "incubator",
					isInstalled: true,
				},
			},
			nil,
		)
		pmMock.On("RemotePackages").Return(
			[]pkg.Package{
				myPkg{
					name:        "lib2",
					version:     "master",
					registry:    "incubator",
					isInstalled: false,
				},
			},
			nil,
		)
		pmMock.On("PackageEnvironments", mock.Anything).Return(
			func(p pkg.Package) []*app.EnvironmentConfig {
				switch {
				case p.Name() == "lib1" && p.Version() == "0.0.1":
					return []*app.EnvironmentConfig{
						&app.EnvironmentConfig{
							Name: "production",
						},
						&app.EnvironmentConfig{
							Name: "default",
						},
					}
				case p.Name() == "lib1" && p.Version() == "0.0.2":
					return []*app.EnvironmentConfig{
						&app.EnvironmentConfig{
							Name: "stage",
						},
					}
				default:
					return nil
				}
			}, nil,
		)

		cases := []struct {
			name          string
			onlyInstalled bool
			outputType    string
			outputName    string
			isErr         bool
		}{
			{
				name:       "list all packages",
				outputName: "pkg/list/output.txt",
			},
			{
				name:       "output json",
				outputType: "json",
				outputName: "pkg/list/output.json",
			},
			{
				name:          "installed packages",
				onlyInstalled: true,
				outputName:    "pkg/list/installed.txt",
			},
			{
				name:       "invalid output type",
				outputType: "invalid",
				isErr:      true,
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				in := map[string]interface{}{
					OptionApp:           appMock,
					OptionInstalled:     tc.onlyInstalled,
					OptionOutput:        tc.outputType,
					OptionTLSSkipVerify: false,
				}

				a, err := NewPkgList(in)
				require.NoError(t, err)
				a.pm = &pmMock

				var buf bytes.Buffer
				a.out = &buf

				err = a.Run()
				if tc.isErr {
					require.Error(t, err)
					return
				}
				require.NoError(t, err)

				assertOutput(t, tc.outputName, buf.String())
			})
		}

	})
}

func TestPkgList_requires_app(t *testing.T) {
	in := make(map[string]interface{})
	_, err := NewPkgList(in)
	require.Error(t, err)
}

func makePkg(name string) pkg.Package {
	var pkg pmocks.Package
	pkg.On("Name").Return(name)
	return &pkg
}
func TestPkgList_envListForPackage(t *testing.T) {
	var pkgEnvs = map[string][]string{
		"lib1": []string{"z", "y", "x"},
		"lib2": []string{},
	}
	pm := rmocks.PackageManager{}
	pm.On("PackageEnvironments", mock.Anything).Return(
		func(p pkg.Package) []*app.EnvironmentConfig {
			names, ok := pkgEnvs[p.Name()]
			if !ok {
				return nil
			}
			result := make([]*app.EnvironmentConfig, 0, len(names))
			for _, name := range names {
				result = append(result,
					&app.EnvironmentConfig{
						Name: name,
					},
				)
			}
			return result
		},
		func(p pkg.Package) error {
			if _, ok := pkgEnvs[p.Name()]; !ok {
				return errors.New("no such package")
			}
			return nil
		},
	)

	tests := []struct {
		name      string
		pkg       pkg.Package
		expected  string
		expectErr bool
	}{
		{
			name:     "existing package",
			pkg:      makePkg("lib1"),
			expected: "x, y, z",
		},
		{
			name:     "package has not environments",
			pkg:      makePkg("lib2"),
			expected: "",
		},
		{
			name:      "nonexistent  package",
			pkg:       makePkg("no such package"),
			expectErr: true,
		},
	}

	for _, tc := range tests {
		actual, err := envListForPackage(&pm, tc.pkg)
		if tc.expectErr {
			require.Error(t, err, tc.name)
			continue
		}
		require.NoError(t, err, tc.name)

		assert.Equal(t, tc.expected, actual, tc.name)
	}
}
