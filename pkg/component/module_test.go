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

package component

import (
	"path/filepath"
	"testing"

	"github.com/ksonnet/ksonnet/pkg/app"
	"github.com/ksonnet/ksonnet/pkg/app/mocks"
	"github.com/ksonnet/ksonnet/pkg/util/test"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_GetModule(t *testing.T) {
	cases := []struct {
		name         string
		moduleName   string
		dir          string
		isErr        bool
		expectedName string
	}{
		{
			name:         "root module",
			moduleName:   "/",
			dir:          "/app/components",
			expectedName: "/",
		},
		{
			name:         "empty name",
			moduleName:   "",
			dir:          "/app/components",
			expectedName: "/",
		},
		{
			name:         "nested",
			moduleName:   "nested",
			dir:          "/app/components/nested",
			expectedName: "nested",
		},
		{
			name:         "nested deeply",
			moduleName:   "deep.nested",
			dir:          "/app/components/deep/nested",
			expectedName: "deep.nested",
		},
		{
			name:       "path doesn't exist",
			moduleName: "invalid",
			isErr:      true,
		},
		{
			name:       "invalid name",
			moduleName: "!!",
			isErr:      true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			test.WithApp(t, "/app", func(a *mocks.App, fs afero.Fs) {
				if tc.dir != "" {
					err := fs.MkdirAll(tc.dir, 0755)
					afero.WriteFile(fs, filepath.Join(tc.dir, paramsFile), []byte("{}"), app.DefaultFolderPermissions)
					require.NoError(t, err)
				}

				m, err := GetModule(a, tc.moduleName)
				if tc.isErr {
					require.Error(t, err)
					return
				}

				require.NoError(t, err)

				assert.Equal(t, tc.expectedName, m.Name())
			})
		})
	}
}

func TestModule_Components(t *testing.T) {
	test.WithApp(t, "/app", func(a *mocks.App, fs afero.Fs) {
		test.StageFile(t, fs, "certificate-crd.yaml", "/app/components/module1/certificate-crd.yaml")
		test.StageFile(t, fs, "params-with-entry.libsonnet", "/app/components/module1/params.libsonnet")
		test.StageFile(t, fs, "params-no-entry.libsonnet", "/app/components/params.libsonnet")

		cases := []struct {
			name   string
			module string
			count  int
		}{
			{
				name:   "no components",
				module: "/",
			},
			{
				name:   "with components",
				module: "module1",
				count:  1,
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {

				ns, err := GetModule(a, tc.module)
				require.NoError(t, err)

				assert.Equal(t, tc.module, ns.Name())
				components, err := ns.Components()
				require.NoError(t, err)

				assert.Len(t, components, tc.count)
			})
		}
	})
}

func TestFilesystemModule_DeleteParam(t *testing.T) {
	test.WithApp(t, "/app", func(a *mocks.App, fs afero.Fs) {
		test.StageFile(t, fs, "params-global.libsonnet", "/app/components/params.libsonnet")

		module := NewModule(a, ".")

		err := module.DeleteParam([]string{"metadata"})
		require.NoError(t, err)

		test.AssertContents(t, fs, "params-delete-global.libsonnet", "/app/components/params.libsonnet")
	})
}

func TestExtractModuleComponent(t *testing.T) {
	cases := []struct {
		name string
		in   string
		c    string
		m    string
	}{
		{
			name: "no module",
			in:   "component",
			c:    "component",
			m:    "/",
		},
		{
			name: "with module",
			in:   "module/component",
			c:    "component",
			m:    "module",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			test.WithApp(t, "/app", func(a *mocks.App, fs afero.Fs) {
				m, c := ExtractModuleComponent(a, tc.in)

				assert.Equal(t, tc.m, m.Name())
				assert.Equal(t, tc.c, c)
			})
		})
	}
}

func TestFromName(t *testing.T) {
	cases := []struct {
		name string
		in   string
		c    string
		m    string
	}{
		{
			name: "no module",
			in:   "component",
			c:    "component",
			m:    "",
		},
		{
			name: "with module",
			in:   "module.component",
			c:    "component",
			m:    "module",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m, c := FromName(tc.in)

			assert.Equal(t, tc.m, m)
			assert.Equal(t, tc.c, c)
		})
	}
}

func TestModuleFromPath(t *testing.T) {
	cases := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "path in application",
			path:     "/app/components",
			expected: "",
		},
		{
			name:     "nested component",
			path:     "/app/components/nested",
			expected: "nested",
		},
		{
			name:     "deeply nested component",
			path:     "/app/components/nested/deeply",
			expected: "nested.deeply",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			test.WithApp(t, "/app", func(a *mocks.App, fs afero.Fs) {
				modulePath := ModuleFromPath(a, tc.path)
				require.Equal(t, tc.expected, modulePath)
			})
		})
	}
}
