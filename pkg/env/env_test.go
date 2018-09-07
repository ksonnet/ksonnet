// Copyright 2018 The kubecfg authors
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

package env

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/ksonnet/ksonnet/pkg/app"
	"github.com/ksonnet/ksonnet/pkg/app/mocks"
	"github.com/ksonnet/ksonnet/pkg/pkg"
	pmocks "github.com/ksonnet/ksonnet/pkg/pkg/mocks"
	"github.com/ksonnet/ksonnet/pkg/registry"
	rmocks "github.com/ksonnet/ksonnet/pkg/registry/mocks"
	"github.com/ksonnet/ksonnet/pkg/util/jsonnet"
	"github.com/ksonnet/ksonnet/pkg/util/test"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddJPaths(t *testing.T) {
	withJsonnetPaths(func() {
		AddJPaths("/vendor")
		require.Equal(t, []string{"/vendor"}, componentJPaths)
	})
}

func TestJPathEmpty(t *testing.T) {
	require.Empty(t, componentJPaths)
}

func TestAddExtVar(t *testing.T) {
	type args struct {
		key   string
		value string
	}
	testCases := []struct {
		name string
		args args
	}{
		{
			name: "add a key and value",
			args: args{
				key:   "key",
				value: "value",
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			withJsonnetPaths(func() {
				AddExtVar(tc.args.key, tc.args.value)
				require.Equal(t, tc.args.value, componentExtVars[tc.args.key])
			})
		})
	}
}

func TestAddExtVarFile(t *testing.T) {
	type args struct {
		key  string
		file string
	}
	testCases := []struct {
		name          string
		args          args
		expectedValue string
		stagePath     string
		isErr         bool
	}{
		{
			name: "add a key and value",
			args: args{
				key:  "key",
				file: "/app/value.txt",
			},
			expectedValue: "value",
			stagePath:     "/app/value.txt",
		},
		{
			name: "add a key and value",
			args: args{
				key:  "key",
				file: "/app/value.txt",
			},
			isErr: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			test.WithApp(t, "/app", func(a *mocks.App, fs afero.Fs) {
				withJsonnetPaths(func() {
					if tc.stagePath != "" {
						test.StageFile(t, fs, "value.txt", tc.stagePath)
					}
					err := AddExtVarFile(fs, tc.args.key, tc.args.file)
					if tc.isErr {
						require.Error(t, err)
						return
					}
					require.NoError(t, err)
					require.Equal(t, tc.expectedValue, componentExtVars[tc.args.key])

				})
			})
		})
	}
}

func TestAddTlaVar(t *testing.T) {
	type args struct {
		key   string
		value string
	}
	testCases := []struct {
		name          string
		args          args
		expectedKey   string
		expectedValue string
	}{
		{
			name: "add a key and value",
			args: args{
				key:   "key",
				value: "value",
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			withJsonnetPaths(func() {
				AddTlaVar(tc.args.key, tc.args.value)
				require.Equal(t, tc.args.value, componentTlaVars[tc.args.key])
			})
		})
	}
}

func TestAddTlaVarFile(t *testing.T) {
	type args struct {
		key  string
		file string
	}
	testCases := []struct {
		name          string
		args          args
		expectedValue string
		stagePath     string
		isErr         bool
	}{
		{
			name: "add a key and value",
			args: args{
				key:  "key",
				file: "/app/value.txt",
			},
			expectedValue: "value",
			stagePath:     "/app/value.txt",
		},
		{
			name: "add a key and value",
			args: args{
				key:  "key",
				file: "/app/value.txt",
			},
			isErr: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			test.WithApp(t, "/app", func(a *mocks.App, fs afero.Fs) {
				withJsonnetPaths(func() {
					if tc.stagePath != "" {
						test.StageFile(t, fs, "value.txt", tc.stagePath)
					}
					err := AddTlaVarFile(fs, tc.args.key, tc.args.file)
					if tc.isErr {
						require.Error(t, err)
						return
					}
					require.NoError(t, err)
					require.Equal(t, tc.expectedValue, componentTlaVars[tc.args.key])

				})
			})
		})
	}
}

func withJsonnetPaths(fn func()) {
	ogComponentJPaths := componentJPaths
	ogComponentExtVars := componentExtVars
	ogComponentTlaVars := componentTlaVars

	defer func() {
		componentJPaths = ogComponentJPaths
		componentExtVars = ogComponentExtVars
		componentTlaVars = ogComponentTlaVars
	}()

	fn()
}

func TestEvaluate(t *testing.T) {
	test.WithApp(t, "/app", func(a *mocks.App, fs afero.Fs) {
		envSpec := &app.EnvironmentConfig{
			Path: "default",
			Destination: &app.EnvironmentDestinationSpec{
				Server:    "http://example.com",
				Namespace: "default",
			},
		}
		a.On("Environment", "default").Return(envSpec, nil)
		a.On("Libraries").Return(app.LibraryConfigs{}, nil)
		a.On("Registries").Return(app.RegistryConfigs{}, nil)

		test.StageFile(t, fs, "main.jsonnet", "/app/environments/default/main.jsonnet")

		components, err := ioutil.ReadFile(filepath.FromSlash("testdata/evaluate/components.jsonnet"))
		require.NoError(t, err)

		got, err := Evaluate(a, "default", string(components), "")
		require.NoError(t, err)

		test.AssertOutput(t, "evaluate/out.jsonnet", got)
	})
}

func TestEvaluate_versionedPackages(t *testing.T) {
	require.Empty(t, componentJPaths)

	test.WithApp(t, "/app", func(a *mocks.App, fs afero.Fs) {
		envSpec := &app.EnvironmentConfig{
			Path: "default",
			Destination: &app.EnvironmentDestinationSpec{
				Server:    "http://example.com",
				Namespace: "default",
			},
			Libraries: app.LibraryConfigs{
				"printer": &app.LibraryConfig{
					Name:     "printer",
					Registry: "incubator",
					Version:  "0.0.1",
				},
			},
		}
		a.On("Environment", "default").Return(envSpec, nil)
		a.On("Libraries").Return(app.LibraryConfigs{}, nil)
		a.On("Registries").Return(app.RegistryConfigs{
			"incubator": &app.RegistryConfig{
				Name:     "incubator",
				Protocol: string(registry.ProtocolFilesystem),
			},
		}, nil)
		a.On("VendorPath").Return("/app/vendor")

		// Stage environment, packages
		test.StageDir(t, fs, "versionedPackageApp/vendor", "/app/vendor")
		test.StageDir(t, fs, "versionedPackageApp/environments", "/app/environments")
		test.DumpFs(t, fs) // DELETEME

		components, err := ioutil.ReadFile(filepath.FromSlash("testdata/versionedPackageApp/components/hello-world.jsonnet"))
		require.NoError(t, err)

		got, err := Evaluate(a, "default", string(components), "", jsonnet.AferoImporterOpt(fs))
		require.NoError(t, err)

		test.AssertOutput(t, "versionedPackageApp/expected.json", got)
	})
}

func TestMainFile(t *testing.T) {
	test.WithApp(t, "/app", func(a *mocks.App, fs afero.Fs) {
		envSpec := &app.EnvironmentConfig{}
		a.On("Environment", "default").Return(envSpec, nil)

		test.StageFile(t, fs, "main.jsonnet", "/app/environments/main.jsonnet")

		got, err := MainFile(a, "default")
		require.NoError(t, err)

		test.AssertOutput(t, "main.jsonnet", got)
	})
}

func Test_upgradeArray(t *testing.T) {
	snippet, err := ioutil.ReadFile(filepath.FromSlash("testdata/upgradeArray/in.jsonnet"))
	require.NoError(t, err)

	got, err := upgradeArray(string(snippet))
	require.NoError(t, err)

	test.AssertOutput(t, "upgradeArray/out.jsonnet", got)
}

// Helper for creating mock pkg.Package
func makePackage(registry string, name string, version string, installed bool) pkg.Package {
	p := new(pmocks.Package)
	p.On("Name").Return(name)
	p.On("RegistryName").Return(registry)
	p.On("Version").Return(version)
	p.On("IsInstalled").Return(installed)
	p.On("Path").Return(
		filepath.Join("/", "app", "vendor", registry, fmt.Sprintf("%s@%s", name, version)),
	)
	p.On("String").Return(
		fmt.Sprintf("%s/%s@%s", registry, name, version),
	)
	return p
}

func Test_buildPackagePaths(t *testing.T) {
	// Rig a package manager to return a fixed set of packages for the environment
	r := "incubator"
	e := &app.EnvironmentConfig{Name: "default"}
	pkgByName := map[string]pkg.Package{
		"incubator/nginx":       makePackage(r, "nginx", "1.2.3", true),
		"incubator/mysql":       makePackage(r, "mysql", "00112233ff", true),
		"incubator/unversioned": makePackage(r, "unversioned", "", true),
	}
	packages := make([]pkg.Package, 0, len(pkgByName))
	for _, p := range pkgByName {
		packages = append(packages, p)
	}
	pm := new(rmocks.PackageManager)
	pm.On("PackagesForEnv", e).Return(packages, nil)

	results, err := buildPackagePaths(pm, e)
	require.NoError(t, err)

	// Ensure all expected packages are in results and their paths match.
	for name, path := range results {
		p, ok := pkgByName[name]
		assert.True(t, ok, "unexpected package: %v", name)
		if p != nil {
			assert.Equal(t, path, p.Path(), "package %v vendor path mismatch", name)
		}
	}

	// Ensure all versioned packages are in the results.
	for name, p := range pkgByName {
		if p.Version() == "" {
			// Unversioned packages are expected to have been filtered out of the results.
			continue
		}
		assert.Contains(t, results, name, "expected package not in results")
	}

}

func Test_revendorPackages(t *testing.T) {
	// Rig a package manager to return a fixed set of packages for the environment
	r := "incubator"
	e := &app.EnvironmentConfig{Name: "default"}
	pkgByName := map[string]pkg.Package{
		"nginx":     makePackage(r, "nginx", "1.2.3", true),
		"mysql":     makePackage(r, "mysql", "00112233ff", true),
		"not_there": makePackage(r, "not_there", "2.3.4", true),
	}
	packages := make([]pkg.Package, 0, len(pkgByName))
	for _, p := range pkgByName {
		packages = append(packages, p)
	}
	pm := new(rmocks.PackageManager)
	pm.On("PackagesForEnv", e).Return(packages, nil)

	test.WithApp(t, "/", func(a *mocks.App, fs afero.Fs) {
		// Stage some packages to copy
		for _, p := range packages {
			srcPath := filepath.Join("packages", p.Name())
			if _, err := os.Stat(filepath.Join("testdata", srcPath)); os.IsNotExist(err) {
				// Skip staging missing packages
				continue
			}
			test.StageDir(t, fs, srcPath, p.Path())
		}
		test.DumpFs(t, fs)

		newRoot, cleanup, err := revendorPackages(a, pm, e)
		defer func() {
			if cleanup == nil {
				return
			}

			cleanup()
			exists, err := afero.DirExists(fs, newRoot)
			assert.NoError(t, err, "checking cleanup")
			assert.NotEqual(t, true, exists, "cleanup func did not remove directory: %v", newRoot)
		}()
		require.NoError(t, err, "revendoring packages")
		require.NotEmpty(t, newRoot, "vendored path")

		// TODO check tempPath is not contained in original vendor path

		// Verify structure of copied packages (sans-version)
		for _, p := range packages {
			// Our assumption is that packages have been re-homed from their original
			// path: `vendor/<registry>/<pkg>@<version>`
			// to their unversioned temporary path:
			// `<newRoot>/<registry>/<pkg>`
			oldPath := p.Path()
			newPath := filepath.Join(newRoot, p.RegistryName(), p.Name())

			if ok, err := afero.Exists(fs, oldPath); !ok || err != nil {
				t.Logf("skipping package %v, it was not staged", p)
				continue
			}
			t.Logf("comparing paths %v and %v...", oldPath, newPath)
			test.AssertDirectoriesMatch(t, fs, oldPath, newPath)
		}

	})
}
