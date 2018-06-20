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

package app

import (
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_baseapp_CurrentEnvironment(t *testing.T) {
	cases := []struct {
		name     string
		init     func(*testing.T, afero.Fs)
		expected string
	}{
		{
			name: "without a current environment set",
		},
		{
			name: "with a current environment set",
			init: func(t *testing.T, fs afero.Fs) {
				path := filepath.Join("/", currentEnvName)
				err := afero.WriteFile(fs, path, []byte("default"), DefaultFilePermissions)
				require.NoError(t, err)
			},
			expected: "default",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fs := afero.NewMemMapFs()
			stageFile(t, fs, "app010_app.yaml", "/app.yaml")
			ba := newBaseApp(fs, "/")

			if tc.init != nil {
				tc.init(t, fs)
			}
			got := ba.CurrentEnvironment()
			require.Equal(t, tc.expected, got)
		})
	}
}

func Test_baseapp_SetCurrentEnvironment(t *testing.T) {
	fs := afero.NewMemMapFs()
	stageFile(t, fs, "app010_app.yaml", "/app.yaml")
	ba := newBaseApp(fs, "/")

	err := ba.SetCurrentEnvironment("default")
	require.NoError(t, err)

	current := ba.CurrentEnvironment()
	assert.Equal(t, "default", current)
}

func Test_baseApp_AddRegistry(t *testing.T) {
	fs := afero.NewMemMapFs()

	stageFile(t, fs, "app010_app.yaml", "/app.yaml")

	ba := newBaseApp(fs, "/")

	reg := &RegistryConfig{
		Name: "new",
	}

	err := ba.AddRegistry(reg, false)
	require.NoError(t, err)

	assertContents(t, fs, "add-registry.yaml", ba.configPath())
	assertNotExists(t, fs, ba.overridePath())
}

func Test_baseApp_AddRegistry_override(t *testing.T) {
	fs := afero.NewMemMapFs()

	stageFile(t, fs, "app010_app.yaml", "/app.yaml")

	ba := newBaseApp(fs, "/")

	reg := &RegistryConfig{
		Name: "new",
	}

	err := ba.AddRegistry(reg, true)
	require.NoError(t, err)

	assertContents(t, fs, "app020_app.yaml", ba.configPath())
	assertContents(t, fs, "add-registry-override.yaml", ba.overridePath())
}

func Test_baseApp_AddRegistry_override_existing(t *testing.T) {
	fs := afero.NewMemMapFs()

	stageFile(t, fs, "app010_app.yaml", "/app.yaml")

	ba := newBaseApp(fs, "/")

	reg := &RegistryConfig{
		Name: "incubator",
	}

	err := ba.AddRegistry(reg, true)
	require.NoError(t, err)
}

func Test_baseApp_UpdateRegistry(t *testing.T) {
	tests := []struct {
		name           string
		regSpec        RegistryConfig
		appFilePath    string
		expectFilePath string
		expectErr      bool
	}{
		{
			name:           "no such registry",
			regSpec:        RegistryConfig{Name: "no-such-registry"},
			appFilePath:    "app010_app.yaml",
			expectFilePath: "",
			expectErr:      true,
		},
	}

	for _, tc := range tests {
		fs := afero.NewMemMapFs()

		if tc.appFilePath != "" {
			stageFile(t, fs, tc.appFilePath, "/app.yaml")
		}

		ba := newBaseApp(fs, "/")

		// Test updating non-existing registry
		err := ba.UpdateRegistry(&tc.regSpec)
		if tc.expectErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
		}

		if tc.expectFilePath != "" {
			assertContents(t, fs, tc.expectFilePath, ba.configPath())
		}
		//assertNotExists(t, fs, ba.overridePath())
	}

}

func Test_baseApp_load_override(t *testing.T) {
	fs := afero.NewMemMapFs()

	stageFile(t, fs, "app010_app.yaml", "/app.yaml")
	stageFile(t, fs, "add-registry-override.yaml", "/app.override.yaml")

	ba := newBaseApp(fs, "/")

	err := ba.load()
	require.NoError(t, err)

	_, ok := ba.overrides.Registries["new"]
	require.True(t, ok)
}

func Test_baseApp_load_override_invalid(t *testing.T) {
	fs := afero.NewMemMapFs()

	stageFile(t, fs, "app010_app.yaml", "/app.yaml")
	stageFile(t, fs, "add-registry-override-invalid.yaml", "/app.override.yaml")

	ba := newBaseApp(fs, "/")

	err := ba.load()
	require.Error(t, err)
}
