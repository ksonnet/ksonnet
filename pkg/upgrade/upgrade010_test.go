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

package upgrade

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/ksonnet/ksonnet/pkg/app"
	amocks "github.com/ksonnet/ksonnet/pkg/app/mocks"
	"github.com/ksonnet/ksonnet/pkg/pkg"
	pmocks "github.com/ksonnet/ksonnet/pkg/pkg/mocks"
	rmocks "github.com/ksonnet/ksonnet/pkg/registry/mocks"
	"github.com/ksonnet/ksonnet/pkg/util/test"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type fakeLibUpdater func(k8sSpecFlag string, libPath string) (string, error)

func (f fakeLibUpdater) UpdateKSLib(k8sSpecFlag string, libPath string) (string, error) {
	return f(k8sSpecFlag, libPath)
}

func withApp010Fs(t *testing.T, appName string, fn func(app app.App)) {
	fs := afero.NewMemMapFs()

	envDirs := []string{
		"default",
		"us-east/test",
		"us-west/test",
		"us-west/prod",
	}

	for _, dir := range envDirs {
		path := filepath.Join("/environments", dir)
		err := fs.MkdirAll(path, app.DefaultFolderPermissions)
		require.NoError(t, err)

		swaggerPath := filepath.Join(path, "main.jsonnet")
		test.StageFile(t, fs, "main.jsonnet", swaggerPath)
	}

	test.StageFile(t, fs, appName, "/app.yaml")

	libUpdaterOpt := app.OptLibUpdater(
		fakeLibUpdater(func(k8sSpecFlag string, libPath string) (string, error) {
			return "v1.8.7", nil
		}),
	)
	a := app.NewBaseApp(fs, "/", nil, libUpdaterOpt)

	fn(a)
}

func TestApp010_Upgrade(t *testing.T) {
	var b bytes.Buffer
	var defaultPM rmocks.PackageManager
	defaultPM.On("Packages").Return([]pkg.Package{}, nil)

	cases := []struct {
		name         string
		stageFile    string
		init         func(t *testing.T, a app.App) upgrader
		checkUpgrade func(t *testing.T, a app.App)
		dryRun       bool
		isErr        bool
	}{
		{
			name:      "ksonnet lib doesn't need to be upgraded",
			stageFile: "app010_app.yaml",
			init: func(t *testing.T, a app.App) upgrader {
				err := a.Fs().MkdirAll("/lib", app.DefaultFolderPermissions)
				require.NoError(t, err)

				p := filepath.Join(a.Root(), "lib", "ksonnet-lib", "v1.10.3")
				err = a.Fs().MkdirAll(p, app.DefaultFolderPermissions)
				require.NoError(t, err)

				return newUpgrade010(a, &b, &defaultPM)
			},
			dryRun: false,
		},
		{
			name:      "ksonnet lib needs to be upgraded",
			stageFile: "app010_app.yaml",
			init: func(t *testing.T, a app.App) upgrader {
				err := a.Fs().MkdirAll("/lib", app.DefaultFolderPermissions)
				require.NoError(t, err)

				p := filepath.Join(a.Root(), "lib", "v1.10.3")
				err = a.Fs().MkdirAll(p, app.DefaultFolderPermissions)
				require.NoError(t, err)

				return newUpgrade010(a, &b, &defaultPM)
			},
			dryRun: false,
		},
		{
			name:      "ksonnet lib needs to be upgraded - dry run",
			stageFile: "app010_app.yaml",
			init: func(t *testing.T, a app.App) upgrader {
				err := a.Fs().MkdirAll("/lib", app.DefaultFolderPermissions)
				require.NoError(t, err)

				p := filepath.Join(a.Root(), "lib", "v1.10.3")
				err = a.Fs().MkdirAll(p, app.DefaultFolderPermissions)
				require.NoError(t, err)

				return newUpgrade010(a, &b, &defaultPM)
			},
			checkUpgrade: func(t *testing.T, a app.App) {
				isDir, err := afero.IsDir(a.Fs(), filepath.Join("/lib", "v1.10.3"))
				require.NoError(t, err)
				require.True(t, isDir)
			},
			dryRun: true,
		},
		{
			name:      "lib doesn't exist",
			stageFile: "app010_app.yaml",
			isErr:     true,
			init: func(t *testing.T, a app.App) upgrader {
				return newUpgrade010(a, &b, &defaultPM)
			},
		},
		{
			name:      "vendored packages need to be upgraded",
			stageFile: "app010_app.yaml",
			init: func(t *testing.T, a app.App) upgrader {
				err := a.Fs().MkdirAll("/lib", app.DefaultFolderPermissions)
				require.NoError(t, err)

				test.StageDir(t, a.Fs(), filepath.FromSlash("packages/mysql"), filepath.FromSlash("/vendor/incubator/mysql"))

				var p pmocks.Package
				var pm rmocks.PackageManager
				p.On("Version").Return("1.2.3")
				p.On("String").Return("incubator/mysql@1.2.3")
				p.On("Path").Return("/vendor/incubator/mysql@1.2.3")
				pm.On("Packages").Return(
					[]pkg.Package{&p}, nil,
				)
				return newUpgrade010(a, &b, &pm)
			},
			checkUpgrade: func(t *testing.T, a app.App) {
				ok, err := afero.DirExists(a.Fs(), "/vendor/incubator/mysql@1.2.3")
				require.NoError(t, err)
				assert.True(t, ok, "checking for upgraded package path")
			},
			dryRun: false,
		},
		{
			name:      "unversioned packages are untouched",
			stageFile: "app010_app.yaml",
			init: func(t *testing.T, a app.App) upgrader {
				err := a.Fs().MkdirAll("/lib", app.DefaultFolderPermissions)
				require.NoError(t, err)

				test.StageDir(t, a.Fs(), filepath.FromSlash("packages/mysql"), filepath.FromSlash("/vendor/incubator/mysql"))

				var p pmocks.Package
				var pm rmocks.PackageManager
				p.On("Version").Return("")
				pm.On("Packages").Return(
					[]pkg.Package{&p}, nil,
				)
				return newUpgrade010(a, &b, &pm)
			},
			checkUpgrade: func(t *testing.T, a app.App) {
				ok, err := afero.DirExists(a.Fs(), "/vendor/incubator/mysql")
				require.NoError(t, err)
				assert.True(t, ok, "checking for upgraded package path")
			},
			dryRun: false,
		},
		{
			name:      "environment targets are upgraded",
			stageFile: "app010_env_targets.yaml",
			init: func(t *testing.T, a app.App) upgrader {
				err := a.Fs().MkdirAll("/lib", app.DefaultFolderPermissions)
				require.NoError(t, err)

				p := filepath.Join(a.Root(), "lib", "ksonnet-lib", "v1.10.3")
				err = a.Fs().MkdirAll(p, app.DefaultFolderPermissions)
				require.NoError(t, err)

				return newUpgrade010(a, &b, &defaultPM)
			},
			checkUpgrade: func(t *testing.T, a app.App) {
				test.AssertContents(t, a.Fs(), "app030_env_targets.yaml", "/app.yaml")
			},
			dryRun: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			withApp010Fs(t, tc.stageFile, func(a app.App) {
				var u upgrader
				if tc.init != nil {
					u = tc.init(t, a)
				} else {
					u = newUpgrade010(a, &b, &defaultPM)
				}

				err := u.Upgrade(tc.dryRun)
				if tc.isErr {
					require.Error(t, err, tc.name)
					return
				}

				require.NoError(t, err, tc.name)

				if tc.checkUpgrade != nil {
					tc.checkUpgrade(t, a)
				}
				b.Reset()
			})
		})
	}
}

func Test_upgrade010_upgradeEnvTargets(t *testing.T) {
	makeApp := func(targets map[string][]string) *amocks.App {
		envs := make(app.EnvironmentConfigs)
		for name, t := range targets {
			tCopy := make([]string, len(t))
			copy(tCopy, t)
			envs[name] = &app.EnvironmentConfig{
				Name:    name,
				Targets: tCopy,
			}
		}

		var a = new(amocks.App)
		a.On("Environments").Return(func() app.EnvironmentConfigs {
			return envs
		}, nil)
		a.On("UpdateTargets", mock.Anything, mock.Anything).Return(func(name string, targets []string) error {
			tCopy := make([]string, len(targets))
			copy(tCopy, targets)
			e, ok := envs[name]
			if !ok {
				return errors.Errorf("unknown environment: %s", name)
			}
			e.Targets = tCopy
			return nil
		})

		return a
	}

	tests := []struct {
		name         string
		targets      map[string][]string
		expected     map[string][]string
		dryRun       bool
		expectUpdate bool
		wantErr      bool
	}{
		{
			name: "nothing to upgrade",
			targets: map[string][]string{
				"default": []string{"/", "module_a", "module_b"},
			},
			expected: map[string][]string{
				"default": []string{"/", "module_a", "module_b"},
			},
			expectUpdate: false,
		},
		{
			name: "need to upgrade",
			targets: map[string][]string{
				"default": []string{"/", "prefix/module_b", "prefix/module/module_b"},
				"stage":   []string{"prefix/module_b", "/"},
			},
			expected: map[string][]string{
				"default": []string{"/", "prefix.module_b", "prefix.module.module_b"},
				"stage":   []string{"prefix.module_b", "/"},
			},
			expectUpdate: true,
		},
		{
			name: "dry run",
			targets: map[string][]string{
				"default": []string{"/", "module/module_a", "prefix/module/module_b"},
			},
			expected: map[string][]string{
				"default": []string{"/", "module/module_a", "prefix/module/module_b"},
			},
			dryRun:       true,
			expectUpdate: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var u = new(upgrade010)
			a := makeApp(tt.targets)

			if err := u.upgradeEnvTargets(a, tt.dryRun); (err != nil) != tt.wantErr {
				t.Errorf("upgrade010.upgradeEnvTargets() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.expectUpdate {
				a.AssertNotCalled(t, "UpdateTargets", mock.Anything, mock.Anything)
				return
			}

			a.AssertExpectations(t)
			updated, err := a.Environments()
			require.NoError(t, err)
			for name, targets := range tt.expected {
				assert.Equal(t, targets, updated[name].Targets, "targets for %s", name)
			}
		})
	}
}
