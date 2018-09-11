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
	"testing"

	"github.com/blang/semver"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_migrateSchema010To020(t *testing.T) {
	var src = &Spec010{
		APIVersion:  "0.1.0",
		Kind:        "ksonnet.io/app",
		Name:        "test-migration",
		Version:     "0.0.1",
		Description: "test migration description",
		Authors: []string{
			"Philip K. Dick",
			"Frank Herbert",
			"Nnedi Okorafor",
		},
		Contributors: ContributorSpecs010{
			&ContributorSpec010{
				Name:  "H.P. Lovecraft",
				Email: "cthulu@theabyss.com",
			},
			&ContributorSpec010{
				Name:  "Neil Gaiman",
				Email: "sandman@comics.com",
			},
		},
		Repository: &RepositorySpec010{
			Type: "git",
			URI:  "https://github.com/ksonnet/mixins",
		},
		Bugs:     "fingers crossed",
		Keywords: []string{"ksonnet", "kubernetes"},
		Registries: RegistryConfigs010{
			"incubator": &RegistryConfig010{
				Name:     "incubator",
				Protocol: "github",
				URI:      "github.com/ksonnet/parts/tree/master/incubator",
			},
			"helm-stable": &RegistryConfig010{
				Name:     "helm-stable",
				Protocol: "helm",
				URI:      "https://kubernetes-charts.storage.googleapis.com",
			},
		},
		Environments: EnvironmentConfigs010{
			"default": &EnvironmentConfig010{
				Name:              "default",
				KubernetesVersion: "v1.10.3",
				Path:              "default-path",
				Destination: &EnvironmentDestinationSpec010{
					Server:    "https://localhost:6443",
					Namespace: "default-namespace",
				},
				Targets: []string{"/", "foo/bar/baz", "simple"},
			},
		},
		Libraries: LibraryConfigs010{
			"redis": &LibraryConfig010{
				Name:     "redis",
				Registry: "helm-stable",
				GitVersion: &GitVersionSpec010{
					RefSpec:   "master",
					CommitSHA: "7.6.5",
				},
			},
			"postgres": &LibraryConfig010{
				Name:     "postgres",
				Registry: "incubator",
				GitVersion: &GitVersionSpec010{
					RefSpec:   "master",
					CommitSHA: "9.1.3",
				},
			},
		},
		License: "Apache License 2.0",
	}

	var expected = &Spec020{
		APIVersion:  "0.2.0",
		Kind:        "ksonnet.io/app",
		Name:        "test-migration",
		Version:     "0.0.1",
		Description: "test migration description",
		Authors: []string{
			"Philip K. Dick",
			"Frank Herbert",
			"Nnedi Okorafor",
		},
		Contributors: ContributorSpecs020{
			&ContributorSpec020{
				Name:  "H.P. Lovecraft",
				Email: "cthulu@theabyss.com",
			},
			&ContributorSpec020{
				Name:  "Neil Gaiman",
				Email: "sandman@comics.com",
			},
		},
		Repository: &RepositorySpec020{
			Type: "git",
			URI:  "https://github.com/ksonnet/mixins",
		},
		Bugs:     "fingers crossed",
		Keywords: []string{"ksonnet", "kubernetes"},
		Registries: RegistryConfigs020{
			"incubator": &RegistryConfig020{
				Name:     "incubator",
				Protocol: "github",
				URI:      "github.com/ksonnet/parts/tree/master/incubator",
			},
			"helm-stable": &RegistryConfig020{
				Name:     "helm-stable",
				Protocol: "helm",
				URI:      "https://kubernetes-charts.storage.googleapis.com",
			},
		},
		Environments: EnvironmentConfigs020{
			"default": &EnvironmentConfig020{
				Name:              "default",
				KubernetesVersion: "v1.10.3",
				Path:              "default-path",
				Destination: &EnvironmentDestinationSpec020{
					Server:    "https://localhost:6443",
					Namespace: "default-namespace",
				},
				Targets:   []string{"/", "foo.bar.baz", "simple"},
				Libraries: LibraryConfigs020{},
			},
		},
		Libraries: LibraryConfigs020{
			"redis": &LibraryConfig020{
				Name:     "redis",
				Registry: "helm-stable",
				Version:  "7.6.5",
			},
			"postgres": &LibraryConfig020{
				Name:     "postgres",
				Registry: "incubator",
				Version:  "9.1.3",
			},
		},
		License: "Apache License 2.0",
	}

	actual, err := migrateSchema010To020(src)
	require.NoError(t, err)
	assert.Equal(t, expected, actual)
}
func Test_migrateSchema020To030(t *testing.T) {
	var src = &Spec020{
		APIVersion:  "0.2.0",
		Kind:        "ksonnet.io/app",
		Name:        "test-migration",
		Version:     "0.0.1",
		Description: "test migration description",
		Authors: []string{
			"Philip K. Dick",
			"Frank Herbert",
			"Nnedi Okorafor",
		},
		Contributors: ContributorSpecs020{
			&ContributorSpec020{
				Name:  "H.P. Lovecraft",
				Email: "cthulu@theabyss.com",
			},
			&ContributorSpec020{
				Name:  "Neil Gaiman",
				Email: "sandman@comics.com",
			},
		},
		Repository: &RepositorySpec020{
			Type: "git",
			URI:  "https://github.com/ksonnet/mixins",
		},
		Bugs:     "fingers crossed",
		Keywords: []string{"ksonnet", "kubernetes"},
		Registries: RegistryConfigs020{
			"incubator": &RegistryConfig020{
				Name:     "incubator",
				Protocol: "github",
				URI:      "github.com/ksonnet/parts/tree/master/incubator",
			},
			"helm-stable": &RegistryConfig020{
				Name:     "helm-stable",
				Protocol: "helm",
				URI:      "https://kubernetes-charts.storage.googleapis.com",
			},
		},
		Environments: EnvironmentConfigs020{
			"default": &EnvironmentConfig020{
				Name:              "default",
				KubernetesVersion: "v1.10.3",
				Path:              "default-path",
				Destination: &EnvironmentDestinationSpec020{
					Server:    "https://localhost:6443",
					Namespace: "default-namespace",
				},
				Targets: []string{"/", "foo.bar.baz", "simple"},
				Libraries: LibraryConfigs020{
					"nginx": &LibraryConfig020{
						Name:     "nginx",
						Registry: "incubator",
						Version:  "1.2.3",
					},
					"mysql": &LibraryConfig020{
						Name:     "mysql",
						Registry: "incubator",
						Version:  "8.0.0",
					},
				},
			},
		},
		Libraries: LibraryConfigs020{
			"redis": &LibraryConfig020{
				Name:     "redis",
				Registry: "helm-stable",
				Version:  "7.6.5",
			},
			"postgres": &LibraryConfig020{
				Name:     "postgres",
				Registry: "incubator",
				Version:  "9.1.3",
			},
		},
		License: "Apache License 2.0",
	}

	var expected = &Spec030{
		APIVersion:  "0.3.0",
		Kind:        "ksonnet.io/app",
		Name:        "test-migration",
		Version:     "0.0.1",
		Description: "test migration description",
		Authors: []string{
			"Philip K. Dick",
			"Frank Herbert",
			"Nnedi Okorafor",
		},
		Contributors: ContributorSpecs030{
			&ContributorSpec030{
				Name:  "H.P. Lovecraft",
				Email: "cthulu@theabyss.com",
			},
			&ContributorSpec030{
				Name:  "Neil Gaiman",
				Email: "sandman@comics.com",
			},
		},
		Repository: &RepositorySpec030{
			Type: "git",
			URI:  "https://github.com/ksonnet/mixins",
		},
		Bugs:     "fingers crossed",
		Keywords: []string{"ksonnet", "kubernetes"},
		Registries: RegistryConfigs030{
			"incubator": &RegistryConfig030{
				Name:     "incubator",
				Protocol: "github",
				URI:      "github.com/ksonnet/parts/tree/master/incubator",
			},
			"helm-stable": &RegistryConfig030{
				Name:     "helm-stable",
				Protocol: "helm",
				URI:      "https://kubernetes-charts.storage.googleapis.com",
			},
		},
		Environments: EnvironmentConfigs030{
			"default": &EnvironmentConfig030{
				Name:              "default",
				KubernetesVersion: "v1.10.3",
				Path:              "default-path",
				Destination: &EnvironmentDestinationSpec030{
					Server:    "https://localhost:6443",
					Namespace: "default-namespace",
				},
				Targets: []string{"/", "foo.bar.baz", "simple"},
				Libraries: LibraryConfigs030{
					"incubator/nginx": &LibraryConfig030{
						Name:     "nginx",
						Registry: "incubator",
						Version:  "1.2.3",
					},
					"incubator/mysql": &LibraryConfig030{
						Name:     "mysql",
						Registry: "incubator",
						Version:  "8.0.0",
					},
				},
			},
		},
		Libraries: LibraryConfigs030{
			"helm-stable/redis": &LibraryConfig030{
				Name:     "redis",
				Registry: "helm-stable",
				Version:  "7.6.5",
			},
			"incubator/postgres": &LibraryConfig030{
				Name:     "postgres",
				Registry: "incubator",
				Version:  "9.1.3",
			},
		},
		License: "Apache License 2.0",
	}

	actual, err := migrateSchema020To030(src)
	require.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func Test_Migrator_Load(t *testing.T) {
	var expected = &Spec030{
		APIVersion:  "0.3.0",
		Kind:        "ksonnet.io/app",
		Name:        "test-migration",
		Version:     "0.0.1",
		Description: "test migration description",
		Authors: []string{
			"Philip K. Dick",
			"Frank Herbert",
			"Nnedi Okorafor",
		},
		Contributors: ContributorSpecs030{
			&ContributorSpec030{
				Name:  "H.P. Lovecraft",
				Email: "cthulu@theabyss.com",
			},
			&ContributorSpec030{
				Name:  "Neil Gaiman",
				Email: "sandman@comics.com",
			},
		},
		Repository: &RepositorySpec030{
			Type: "git",
			URI:  "https://github.com/ksonnet/mixins",
		},
		Bugs:     "fingers crossed",
		Keywords: []string{"ksonnet", "kubernetes"},
		Registries: RegistryConfigs030{
			"incubator": &RegistryConfig030{
				Name:     "incubator",
				Protocol: "github",
				URI:      "github.com/ksonnet/parts/tree/master/incubator",
			},
			"helm-stable": &RegistryConfig030{
				Name:     "helm-stable",
				Protocol: "helm",
				URI:      "https://kubernetes-charts.storage.googleapis.com",
			},
		},
		Environments: EnvironmentConfigs030{
			"default": &EnvironmentConfig030{
				Name:              "default",
				KubernetesVersion: "v1.10.3",
				Path:              "default-path",
				Destination: &EnvironmentDestinationSpec030{
					Server:    "https://localhost:6443",
					Namespace: "default-namespace",
				},
				Targets: []string{"/", "foo.bar.baz", "simple"},
				Libraries: LibraryConfigs030{
					"incubator/nginx": &LibraryConfig030{
						Name:     "nginx",
						Registry: "incubator",
						Version:  "1.2.3",
					},
					"incubator/mysql": &LibraryConfig030{
						Name:     "mysql",
						Registry: "incubator",
						Version:  "8.0.0",
					},
				},
			},
		},
		Libraries: LibraryConfigs030{
			"helm-stable/redis": &LibraryConfig030{
				Name:     "redis",
				Registry: "helm-stable",
				Version:  "7.6.5",
			},
			"incubator/postgres": &LibraryConfig030{
				Name:     "postgres",
				Registry: "incubator",
				Version:  "9.1.3",
			},
		},
		License: "Apache License 2.0",
	}

	var expectedFrom010 = &Spec030{
		APIVersion:  "0.3.0",
		Kind:        "ksonnet.io/app",
		Name:        "test-migration",
		Version:     "0.0.1",
		Description: "test migration description",
		Authors: []string{
			"Philip K. Dick",
			"Frank Herbert",
			"Nnedi Okorafor",
		},
		Contributors: ContributorSpecs030{
			&ContributorSpec030{
				Name:  "H.P. Lovecraft",
				Email: "cthulu@theabyss.com",
			},
			&ContributorSpec030{
				Name:  "Neil Gaiman",
				Email: "sandman@comics.com",
			},
		},
		Repository: &RepositorySpec030{
			Type: "git",
			URI:  "https://github.com/ksonnet/mixins",
		},
		Bugs:     "fingers crossed",
		Keywords: []string{"ksonnet", "kubernetes"},
		Registries: RegistryConfigs030{
			"incubator": &RegistryConfig030{
				Name:     "incubator",
				Protocol: "github",
				URI:      "github.com/ksonnet/parts/tree/master/incubator",
			},
			"helm-stable": &RegistryConfig030{
				Name:     "helm-stable",
				Protocol: "helm",
				URI:      "https://kubernetes-charts.storage.googleapis.com",
			},
		},
		Environments: EnvironmentConfigs030{
			"default": &EnvironmentConfig030{
				Name:              "default",
				KubernetesVersion: "v1.10.3",
				Path:              "default-path",
				Destination: &EnvironmentDestinationSpec030{
					Server:    "https://localhost:6443",
					Namespace: "default-namespace",
				},
				Targets:   []string{"/", "foo.bar.baz", "simple"},
				Libraries: LibraryConfigs030{},
			},
		},
		Libraries: LibraryConfigs030{
			"helm-stable/redis": &LibraryConfig030{
				Name:     "redis",
				Registry: "helm-stable",
				Version:  "7.6.5",
			},
			"incubator/postgres": &LibraryConfig030{
				Name:     "postgres",
				Registry: "incubator",
				Version:  "9.1.3",
			},
		},
		License: "Apache License 2.0",
	}
	tests := []struct {
		fromVersion semver.Version
		fromFile    string
		expected    *Spec
	}{
		{
			fromVersion: semver.MustParse("0.1.0"),
			fromFile:    "migrations-app010.yaml",
			expected:    expectedFrom010,
		},
		{
			fromVersion: semver.MustParse("0.2.0"),
			fromFile:    "migrations-app020.yaml",
			expected:    expected,
		},
		{
			fromVersion: semver.MustParse("0.3.0"),
			fromFile:    "migrations-app030.yaml",
			expected:    expected,
		},
	}

	for _, tc := range tests {
		fs := afero.NewMemMapFs()
		stageFile(t, fs, tc.fromFile, "/app.yaml")
		m := NewMigrator(fs, "/")

		actual, err := m.Load(tc.fromVersion, false)
		require.NoErrorf(t, err, "migrating %s", tc.fromFile)
		assert.Equalf(t, tc.expected, actual, "migrating %s", tc.fromFile)
	}
}
