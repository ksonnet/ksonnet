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
			stageFile(t, fs, "app030_app.yaml", "/app.yaml")
			ba := NewBaseApp(fs, "/", nil)

			if tc.init != nil {
				tc.init(t, fs)
			}
			got := ba.CurrentEnvironment()
			require.Equal(t, tc.expected, got)
		})
	}
}

func Test_baseapp_SetCurrentEnvironment(t *testing.T) {
	cases := []struct {
		name    string
		envName string
		isErr   bool
	}{
		{
			name:    "environment exists",
			envName: "default",
		},
		{
			name:    "environment does not exist",
			envName: "invalid",
			isErr:   true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fs := afero.NewMemMapFs()
			stageFile(t, fs, "app030_app.yaml", "/app.yaml")
			ba := NewBaseApp(fs, "/", nil)

			err := ba.SetCurrentEnvironment(tc.envName)
			if tc.isErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			current := ba.CurrentEnvironment()
			assert.Equal(t, tc.envName, current)
		})
	}

}

func Test_baseApp_AddRegistry(t *testing.T) {
	fs := afero.NewMemMapFs()

	stageFile(t, fs, "app030_app.yaml", "/app.yaml")

	ba := NewBaseApp(fs, "/", nil)

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

	stageFile(t, fs, "app030_app.yaml", "/app.yaml")

	ba := NewBaseApp(fs, "/", nil)

	reg := &RegistryConfig{
		Name: "new",
	}

	err := ba.AddRegistry(reg, true)
	require.NoError(t, err)

	assertContents(t, fs, "app030_app.yaml", ba.configPath())
	assertContents(t, fs, "add-registry-override.yaml", ba.overridePath())
}

func Test_baseApp_AddRegistry_override_existing(t *testing.T) {
	fs := afero.NewMemMapFs()

	stageFile(t, fs, "app030_app.yaml", "/app.yaml")

	ba := NewBaseApp(fs, "/", nil)

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
			appFilePath:    "app030_app.yaml",
			expectFilePath: "",
			expectErr:      true,
		},
	}

	for _, tc := range tests {
		fs := afero.NewMemMapFs()

		if tc.appFilePath != "" {
			stageFile(t, fs, tc.appFilePath, "/app.yaml")
		}

		ba := NewBaseApp(fs, "/", nil)

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
	}

}

func Test_baseApp_UpdateLibrary(t *testing.T) {
	tests := []struct {
		name           string
		libCfg         LibraryConfig
		env            string
		appFilePath    string
		expectFilePath string
		expectErr      bool
	}{
		{
			name: "no such environment",
			libCfg: LibraryConfig{
				Name:     "nginx",
				Registry: "incubator",
				Version:  "1.2.3",
			},
			env:         "no-such-environment",
			appFilePath: "app030_app.yaml",
			expectErr:   true,
		},
		{
			name: "success - global scope",
			libCfg: LibraryConfig{
				Name:     "nginx",
				Registry: "incubator",
				Version:  "1.2.3",
			},
			env:            "",
			appFilePath:    "app030_app.yaml",
			expectFilePath: "pkg-install-global.yaml",
			expectErr:      false,
		},
		{
			name: "success - environment scope",
			libCfg: LibraryConfig{
				Name:     "nginx",
				Registry: "incubator",
				Version:  "1.2.3",
			},
			env:            "default",
			appFilePath:    "app030_app.yaml",
			expectFilePath: "pkg-install-env-scope.yaml",
			expectErr:      false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fs := afero.NewMemMapFs()

			if tc.appFilePath != "" {
				stageFile(t, fs, tc.appFilePath, "/app.yaml")
			}

			ba := NewBaseApp(fs, "/", nil)

			// Test updating non-existing registry
			_, err := ba.UpdateLib(tc.libCfg.Name, tc.env, &tc.libCfg)
			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			if tc.expectFilePath != "" {
				assertContents(t, fs, tc.expectFilePath, ba.configPath())
			}
		})
	}

}

func Test_baseApp_UpdateLib_Remove(t *testing.T) {
	tests := []struct {
		name               string
		globalLibraries    LibraryConfigs
		environments       EnvironmentConfigs
		libName            string
		expectRemovedLib   *LibraryConfig
		env                string
		appFilePath        string
		expectEnvironments EnvironmentConfigs
		expectLibraries    LibraryConfigs
		expectErr          bool
	}{
		{
			name:        "remove - env - exists",
			libName:     "nginx",
			env:         "default",
			appFilePath: "app030_simple.yaml",
			environments: EnvironmentConfigs{
				"default": &EnvironmentConfig{
					Name: "default",
					Libraries: LibraryConfigs{
						"nginx": &LibraryConfig{
							Name:     "nginx",
							Registry: "incubator",
							Version:  "1.2.3",
						},
						"other": &LibraryConfig{
							Name:     "other",
							Registry: "incubator",
							Version:  "1.2.3",
						},
					},
				},
			},
			expectRemovedLib: &LibraryConfig{
				Name:     "nginx",
				Registry: "incubator",
				Version:  "1.2.3",
			},
			expectEnvironments: EnvironmentConfigs{
				"default": &EnvironmentConfig{
					Name: "default",
					Libraries: LibraryConfigs{
						"other": &LibraryConfig{
							Name:     "other",
							Registry: "incubator",
							Version:  "1.2.3",
						},
					},
				},
			},
		},
		{
			name:        "remove - global - exists",
			libName:     "nginx",
			appFilePath: "app030_simple.yaml",
			globalLibraries: LibraryConfigs{
				"nginx": &LibraryConfig{
					Name:     "nginx",
					Registry: "incubator",
					Version:  "1.2.3",
				},
				"other": &LibraryConfig{
					Name:     "other",
					Registry: "incubator",
					Version:  "1.2.3",
				},
			},
			expectRemovedLib: &LibraryConfig{
				Name:     "nginx",
				Registry: "incubator",
				Version:  "1.2.3",
			},
			expectLibraries: LibraryConfigs{
				"other": &LibraryConfig{
					Name:     "other",
					Registry: "incubator",
					Version:  "1.2.3",
				},
			},
			expectErr: false,
		},
		{
			name:        "no such environment",
			libName:     "nginx",
			env:         "no-such-environment",
			appFilePath: "app030_app.yaml",
			expectErr:   true,
		},
		{
			name:        "no such library",
			libName:     "no-such-library",
			appFilePath: "app030_app.yaml",
			globalLibraries: LibraryConfigs{
				"nginx": &LibraryConfig{
					Name:     "nginx",
					Registry: "incubator",
					Version:  "1.2.3",
				},
				"other": &LibraryConfig{
					Name:     "other",
					Registry: "incubator",
					Version:  "1.2.3",
				},
			},
			expectErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fs := afero.NewMemMapFs()

			if tc.appFilePath != "" {
				stageFile(t, fs, tc.appFilePath, "/app.yaml")
			}

			ba := NewBaseApp(fs, "/", nil, optNoopLoader())

			if tc.globalLibraries != nil {
				ba.config.Libraries = tc.globalLibraries
			}
			if tc.environments != nil {
				ba.config.Environments = tc.environments
			}

			removed, err := ba.UpdateLib(tc.libName, tc.env, nil)
			if tc.expectErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			assert.Equal(t, tc.expectRemovedLib, removed)
			assert.Equal(t, tc.expectLibraries, ba.config.Libraries)
			assert.Equal(t, tc.expectEnvironments, ba.config.Environments)
		})
	}
}

func Test_baseApp_load_override(t *testing.T) {
	fs := afero.NewMemMapFs()

	stageFile(t, fs, "app030_app.yaml", "/app.yaml")
	stageFile(t, fs, "add-registry-override.yaml", "/app.override.yaml")

	ba := NewBaseApp(fs, "/", nil)

	err := ba.load()
	require.NoError(t, err)

	_, ok := ba.overrides.Registries["new"]
	require.True(t, ok)
}

func Test_baseApp_load_override_invalid(t *testing.T) {
	fs := afero.NewMemMapFs()

	stageFile(t, fs, "app030_app.yaml", "/app.yaml")
	stageFile(t, fs, "add-registry-override-invalid.yaml", "/app.override.yaml")

	ba := NewBaseApp(fs, "/", nil)

	err := ba.load()
	require.Error(t, err)
}

func Test_baseApp_environment_override_is_merged(t *testing.T) {
	fs := afero.NewMemMapFs()
	ba := NewBaseApp(fs, "/", nil, optNoopLoader())
	ba.config.Environments = EnvironmentConfigs{
		"default": &EnvironmentConfig{
			Name: "default",
			Libraries: LibraryConfigs{
				"nginx": &LibraryConfig{Name: "nginx"},
			},
			KubernetesVersion: "v1.7.0",
			Destination: &EnvironmentDestinationSpec{
				Server:    "http://server.com",
				Namespace: "namespace",
			},
			Path:    "default",
			Targets: []string{"target1", "target2"},
		},
	}
	ba.overrides.Environments["default"] = &EnvironmentConfig{
		Name:              "default",
		KubernetesVersion: "v1.8.0",
		Destination: &EnvironmentDestinationSpec{
			Server:    "http://override.com",
			Namespace: "override",
		},
		Path:    "overrides/path",
		Targets: []string{"override1", "override2"},
	}

	expected := &EnvironmentConfig{
		Name: "default",
		Libraries: LibraryConfigs{
			"nginx": &LibraryConfig{Name: "nginx"},
		},
		KubernetesVersion: "v1.8.0",
		Destination: &EnvironmentDestinationSpec{
			Server:    "http://override.com",
			Namespace: "override",
		},
		Path:       "overrides/path",
		Targets:    []string{"override1", "override2"},
		isOverride: true,
	}

	e, err := ba.Environment("default")
	assert.NoError(t, err, "fetching environment")

	assert.Equal(t, expected, e)
}

func Test_baseApp_environment_just_override(t *testing.T) {
	fs := afero.NewMemMapFs()
	ba := NewBaseApp(fs, "/", nil, optNoopLoader())
	ba.config.Environments = EnvironmentConfigs{}
	ba.overrides.Environments["default"] = &EnvironmentConfig{
		Name:              "default",
		KubernetesVersion: "v1.8.0",
		Destination: &EnvironmentDestinationSpec{
			Server:    "http://override.com",
			Namespace: "override",
		},
		Path:       "overrides/path",
		Targets:    []string{"override1", "override2"},
		isOverride: false,
	}

	expected := &EnvironmentConfig{
		Name:              "default",
		KubernetesVersion: "v1.8.0",
		Destination: &EnvironmentDestinationSpec{
			Server:    "http://override.com",
			Namespace: "override",
		},
		Path:       "overrides/path",
		Targets:    []string{"override1", "override2"},
		isOverride: true,
	}

	e, err := ba.Environment("default")
	assert.NoError(t, err, "fetching environment")

	assert.Equal(t, expected, e)
}
