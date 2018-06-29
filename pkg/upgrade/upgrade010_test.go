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
	"github.com/ksonnet/ksonnet/pkg/pkg"
	pmocks "github.com/ksonnet/ksonnet/pkg/pkg/mocks"
	rmocks "github.com/ksonnet/ksonnet/pkg/registry/mocks"
	"github.com/ksonnet/ksonnet/pkg/util/test"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var LibUpdater = app.LibUpdater

type upgrader interface {
	Upgrade(dryRun bool) error
}

func withApp010Fs(t *testing.T, appName string, fn func(app *app.App010)) {
	ogLibUpdater := LibUpdater
	LibUpdater = func(fs afero.Fs, k8sSpecFlag string, libPath string) (string, error) {
		return "v1.8.7", nil
	}

	defer func() {
		LibUpdater = ogLibUpdater
	}()

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

	app := app.NewApp010(fs, "/")

	fn(app)
}

func TestApp010_Upgrade(t *testing.T) {
	var b bytes.Buffer
	var defaultPM rmocks.PackageManager
	defaultPM.On("Packages").Return([]pkg.Package{}, nil)

	cases := []struct {
		name         string
		init         func(t *testing.T, a *app.App010) upgrader
		checkUpgrade func(t *testing.T, a *app.App010)
		dryRun       bool
		isErr        bool
	}{
		{
			name: "ksonnet lib doesn't need to be upgraded",
			init: func(t *testing.T, a *app.App010) upgrader {
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
			name: "ksonnet lib needs to be upgraded",
			init: func(t *testing.T, a *app.App010) upgrader {
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
			name: "ksonnet lib needs to be upgraded - dry run",
			init: func(t *testing.T, a *app.App010) upgrader {
				err := a.Fs().MkdirAll("/lib", app.DefaultFolderPermissions)
				require.NoError(t, err)

				p := filepath.Join(a.Root(), "lib", "v1.10.3")
				err = a.Fs().MkdirAll(p, app.DefaultFolderPermissions)
				require.NoError(t, err)

				return newUpgrade010(a, &b, &defaultPM)
			},
			checkUpgrade: func(t *testing.T, a *app.App010) {
				isDir, err := afero.IsDir(a.Fs(), filepath.Join("/lib", "v1.10.3"))
				require.NoError(t, err)
				require.True(t, isDir)
			},
			dryRun: true,
		},
		{
			name:  "lib doesn't exist",
			isErr: true,
			init: func(t *testing.T, a *app.App010) upgrader {
				return newUpgrade010(a, &b, &defaultPM)
			},
		},
		{
			name: "vendored packages need to be upgraded",
			init: func(t *testing.T, a *app.App010) upgrader {
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
			checkUpgrade: func(t *testing.T, a *app.App010) {
				ok, err := afero.DirExists(a.Fs(), "/vendor/incubator/mysql@1.2.3")
				require.NoError(t, err)
				assert.True(t, ok, "checking for upgraded package path")
			},
			dryRun: false,
		},
		{
			name: "unversion packages are untouched",
			init: func(t *testing.T, a *app.App010) upgrader {
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
			checkUpgrade: func(t *testing.T, a *app.App010) {
				ok, err := afero.DirExists(a.Fs(), "/vendor/incubator/mysql")
				require.NoError(t, err)
				assert.True(t, ok, "checking for upgraded package path")
			},
			dryRun: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			withApp010Fs(t, "app010_app.yaml", func(a *app.App010) {
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
