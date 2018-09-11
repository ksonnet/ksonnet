// Copyright 2017 The kubecfg authors
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
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/ghodss/yaml"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeSimpleRefSpec(name, protocol, uri, version string) *RegistryConfig {
	return &RegistryConfig{
		Name:     name,
		Protocol: protocol,
		URI:      uri,
	}
}

func makeSimpleEnvironmentSpec(name, namespace, server, k8sVersion string) *EnvironmentConfig {
	return &EnvironmentConfig{
		Name: name,
		Destination: &EnvironmentDestinationSpec{
			Namespace: namespace,
			Server:    server,
		},
		KubernetesVersion: k8sVersion,
	}
}

func TestGetRegistryRefSuccess(t *testing.T) {
	example1 := Spec{
		Registries: RegistryConfigs{
			"simple1": &RegistryConfig{
				URI:      "example.com",
				Protocol: "github",
			},
		},
	}

	r1, ok := example1.RegistryConfig("simple1")
	if r1 == nil || !ok {
		t.Error("Expected registry to contain 'simple1'")
	}

	if r1.URI != "example.com" || r1.Name != "simple1" || r1.Protocol != "github" {
		t.Errorf("Registry did not add correct values:\n%v", r1)
	}
}

func TestGetRegistryRefFailure(t *testing.T) {
	example1 := Spec{
		Registries: RegistryConfigs{
			"simple1": &RegistryConfig{
				URI:      "example.com",
				Protocol: "github",
			},
		},
	}

	r1, ok := example1.RegistryConfig("simple2")
	if r1 != nil || ok {
		t.Error("Expected registry to not contain 'simple2'")
	}
}

func TestAddRegistryRefSuccess(t *testing.T) {
	var example1 = Spec{
		Registries: RegistryConfigs{},
	}

	err := example1.AddRegistryConfig(makeSimpleRefSpec("simple1", "github", "example.com", "master"))
	require.NoError(t, err)

	r1, ok1 := example1.RegistryConfig("simple1")
	assert.True(t, ok1)
	expectedR1 := &RegistryConfig{URI: "example.com", Name: "simple1", Protocol: "github"}
	require.Equal(t, expectedR1, r1)

	r2, ok2 := example1.RegistryConfig("simple1")
	assert.True(t, ok2)
	require.Equal(t, expectedR1, r2)

}

func TestAddRegistryRefFailure(t *testing.T) {
	example1 := Spec{
		Registries: RegistryConfigs{
			"simple1": &RegistryConfig{
				URI:      "example.com",
				Protocol: "github",
			},
		},
	}

	err := example1.AddRegistryConfig(makeSimpleRefSpec("", "github", "example.com", "master"))
	if err != ErrRegistryNameInvalid {
		t.Error("Expected registry to fail to add registry with invalid name")
	}

	err = example1.AddRegistryConfig(makeSimpleRefSpec("simple1", "fakeProtocol", "example.com", "master"))
	if err != ErrRegistryExists {
		t.Error("Expected registry to fail to add registry with duplicate name and different protocol")
	}

	err = example1.AddRegistryConfig(makeSimpleRefSpec("simple1", "github", "fakeUrl", "master"))
	if err != ErrRegistryExists {
		t.Error("Expected registry to fail to add registry with duplicate name and different uri")
	}
}

func TestGetEnvironmentConfigs(t *testing.T) {
	example1 := Spec{
		Environments: EnvironmentConfigs{
			"dev": &EnvironmentConfig{
				Destination: &EnvironmentDestinationSpec{
					Namespace: "default",
					Server:    "http://example.com",
				},
				KubernetesVersion: "1.8.0",
			},
		},
	}

	r1 := example1.GetEnvironmentConfigs()
	if len(r1) != 1 {
		t.Error("Expected environments to contain to be of length 1")
	}

	if r1["dev"].Name != "dev" {
		t.Error("Expected to populate name value")
	}
}

// Verifies that EnvironmentConfig.Name fields are injected at unmarshal time.
func TestEnvironmentConfigHasName(t *testing.T) {
	b := []byte(`
apiVersion: 0.2.0
environments:
  default:
    destination:
      namespace: some-namespace
      server: http://example.com
    k8sVersion: v1.7.0
    path: default
  another:
    destination:
      namespace: some-namespace
      server: http://example.com
    k8sVersion: v1.7.0
    path: default
`)

	var spec Spec
	err := yaml.Unmarshal(b, &spec)
	require.NoError(t, err)

	for k, v := range spec.Environments {
		assert.Equal(t, k, v.Name)
	}
}

func TestGetEnvironmentSpecSuccess(t *testing.T) {
	const (
		env        = "dev"
		namespace  = "default"
		server     = "http://example.com"
		k8sVersion = "1.8.0"
	)

	example1 := Spec{
		Environments: EnvironmentConfigs{
			env: &EnvironmentConfig{
				Destination: &EnvironmentDestinationSpec{
					Namespace: namespace,
					Server:    server,
				},
				KubernetesVersion: k8sVersion,
			},
		},
	}

	r1, ok := example1.GetEnvironmentConfig(env)
	if r1 == nil || !ok {
		t.Errorf("Expected environments to contain '%s'", env)
	}

	if r1.Destination.Namespace != namespace ||
		r1.Destination.Server != server || r1.KubernetesVersion != k8sVersion {
		t.Errorf("Environment did not add correct values:\n%v", r1)
	}
}

func TestGetEnvironmentSpecFailure(t *testing.T) {
	example1 := Spec{
		Environments: EnvironmentConfigs{
			"dev": &EnvironmentConfig{
				Destination: &EnvironmentDestinationSpec{
					Namespace: "default",
					Server:    "http://example.com",
				},
				KubernetesVersion: "1.8.0",
			},
		},
	}

	r1, ok := example1.GetEnvironmentConfig("prod")
	if r1 != nil || ok {
		t.Error("Expected environments to not contain 'prod'")
	}
}

func TestAddEnvironmentSpecSuccess(t *testing.T) {
	const (
		env        = "dev"
		namespace  = "default"
		server     = "http://example.com"
		k8sVersion = "1.8.0"
	)

	example1 := Spec{
		Environments: EnvironmentConfigs{},
	}

	err := example1.AddEnvironmentConfig(makeSimpleEnvironmentSpec(env, namespace, server, k8sVersion))
	if err != nil {
		t.Errorf("Expected environment add to succeed:\n%s", err)
	}

	r1, ok1 := example1.GetEnvironmentConfig(env)
	if !ok1 || r1.Destination.Namespace != namespace ||
		r1.Destination.Server != server || r1.KubernetesVersion != k8sVersion {
		t.Errorf("Environment did not add correct values:\n%v", r1)
	}
}

func TestAddEnvironmentSpecFailure(t *testing.T) {
	const (
		envName1   = "dev"
		envName2   = ""
		namespace  = "default"
		server     = "http://example.com"
		k8sVersion = "1.8.0"
	)

	example1 := Spec{
		Environments: EnvironmentConfigs{
			envName1: &EnvironmentConfig{
				Destination: &EnvironmentDestinationSpec{
					Namespace: namespace,
					Server:    server,
				},
				KubernetesVersion: k8sVersion,
			},
		},
	}

	err := example1.AddEnvironmentConfig(makeSimpleEnvironmentSpec(envName2, namespace, server, k8sVersion))
	if err != ErrEnvironmentNameInvalid {
		t.Error("Expected failure while adding environment with an invalid name")
	}

	err = example1.AddEnvironmentConfig(makeSimpleEnvironmentSpec(envName1, namespace, server, k8sVersion))
	if err != ErrEnvironmentExists {
		t.Error("Expected failure while adding environment with duplicate name")
	}
}

func TestDeleteEnvironmentSpec(t *testing.T) {
	example1 := Spec{
		Environments: EnvironmentConfigs{
			"dev": &EnvironmentConfig{
				Destination: &EnvironmentDestinationSpec{
					Namespace: "default",
					Server:    "http://example.com",
				},
				KubernetesVersion: "1.8.0",
			},
		},
	}

	err := example1.DeleteEnvironmentConfig("dev")
	if err != nil {
		t.Error("Expected to successfully delete an environment in spec")
	}

	if _, ok := example1.GetEnvironmentConfig("dev"); ok {
		t.Error("Expected environment 'dev' to be deleted from spec, but still exists")
	}
}

func TestUpdateEnvironmentSpec(t *testing.T) {
	example1 := Spec{
		Environments: EnvironmentConfigs{
			"dev": &EnvironmentConfig{
				Destination: &EnvironmentDestinationSpec{
					Namespace: "default",
					Server:    "http://example.com",
				},
				KubernetesVersion: "1.8.0",
			},
		},
	}

	example2 := EnvironmentConfig{
		Name: "foo",
		Destination: &EnvironmentDestinationSpec{
			Namespace: "foo",
			Server:    "http://example.com",
		},
		KubernetesVersion: "1.8.0",
	}

	err := example1.UpdateEnvironmentConfig("dev", &example2)
	if err != nil {
		t.Error("Expected to successfully update an environment in spec")
	}

	if _, ok := example1.GetEnvironmentConfig("dev"); ok {
		t.Error("Expected environment 'dev' to be deleted from spec, but still exists")
	}

	if _, ok := example1.GetEnvironmentConfig("foo"); !ok {
		t.Error("Expected environment 'foo' to be created in spec, but does not exists")
	}
}

func Test_write(t *testing.T) {
	fs := afero.NewMemMapFs()

	spec := &Spec{
		APIVersion: "0.3.0",
		Environments: EnvironmentConfigs{
			"a": &EnvironmentConfig{},
		},
		Registries: RegistryConfigs{
			"a": &RegistryConfig{},
		},
	}

	err := write(fs, "/", spec)
	require.NoError(t, err)

	assertExists(t, fs, specPath("/"))
	assertContents(t, fs, "write-app.yaml", specPath("/"))
}

func Test_write_no_override(t *testing.T) {
	fs := afero.NewMemMapFs()

	spec := &Spec{
		APIVersion: "0.3.0",
		Environments: EnvironmentConfigs{
			"a": &EnvironmentConfig{},
		},
		Registries: RegistryConfigs{
			"a": &RegistryConfig{},
		},
	}

	err := write(fs, "/", spec)
	require.NoError(t, err)

	assertExists(t, fs, specPath("/"))
	assertContents(t, fs, "write-app.yaml", specPath("/"))

	assertNotExists(t, fs, overridePath("/"))
}

func Test_read(t *testing.T) {
	fs := afero.NewMemMapFs()

	stageFile(t, fs, "write-app.yaml", specPath("/"))
	stageFile(t, fs, "write-override.yaml", overridePath("/"))

	spec, err := read(fs, "/")
	require.NoError(t, err)

	expected := &Spec{
		APIVersion:   "0.3.0",
		Contributors: ContributorSpecs{},
		Environments: EnvironmentConfigs{
			"a": &EnvironmentConfig{Name: "a"},
		},
		Libraries: LibraryConfigs{},
		Registries: RegistryConfigs{
			"a": &RegistryConfig{Name: "a"},
		},
	}

	require.Equal(t, expected, spec)
}

func TestEnvironmentSpec_MakePath(t *testing.T) {
	rootPath := "/"

	spec := EnvironmentConfig{Path: "default"}

	expected := filepath.Join("/", "environments", "default")
	got := spec.MakePath(rootPath)

	require.Equal(t, expected, got)
}

// Test that RegistryConfigs are properly deserialized, specifically
// their Name fields, which are handler by custom UnmarshalJSON code.
func TestUnmarshalRegistryConfigs(t *testing.T) {
	tests := []struct {
		name               string
		input              string
		expectedRegistries int
	}{
		{
			name:               "0.1.0",
			expectedRegistries: 3,
			input: `
apiVersion: 0.1.0
registries:
  incubator:
    gitVersion:
      commitSha: 40285d8a14f1ac5787e405e1023cf0c07f6aa28c
      refSpec: master
    protocol: github
    uri: github.com/ksonnet/parts/tree/master/incubator
  helm-stable:
    protocol: helm
    uri: https://kubernetes-charts.storage.googleapis.com
  otherRegistry:
    protocol: github
    uri: github.com/ksonnet/parts/tree/next/incubator
version: 0.0.1
`,
		},
		{
			name:               "0.2.0",
			expectedRegistries: 3,
			input: `
apiVersion: 0.2.0
registries:
  incubator:
    protocol: github
    uri: github.com/ksonnet/parts/tree/master/incubator
  helm-stable:
    protocol: helm
    uri: https://kubernetes-charts.storage.googleapis.com
  otherRegistry:
    protocol: github
    uri: github.com/ksonnet/parts/tree/next/incubator
version: 0.0.1
`,
		},
	}

	for _, tc := range tests {
		var spec Spec

		err := yaml.Unmarshal([]byte(tc.input), &spec)
		assert.NoError(t, err, tc.name)
		assert.Equal(t, tc.expectedRegistries, len(spec.Registries), tc.name)

		for k, v := range spec.Registries {
			assert.Equal(t, k, v.Name, tc.name)
		}
	}
}

func assertExists(t *testing.T, fs afero.Fs, path string) {
	exists, err := afero.Exists(fs, path)
	require.NoError(t, err)

	require.True(t, exists, "%q does not exist", path)
}

func assertNotExists(t *testing.T, fs afero.Fs, path string) {
	exists, err := afero.Exists(fs, path)
	require.NoError(t, err)

	require.False(t, exists, "%q exists", path)
}

func assertContents(t *testing.T, fs afero.Fs, expectedPath, contentPath string) {
	expected, err := ioutil.ReadFile(filepath.Join("testdata", expectedPath))
	require.NoError(t, err)

	got, err := afero.ReadFile(fs, contentPath)
	require.NoError(t, err)

	require.Equal(t, string(expected), string(got), "unexpected %q contents", contentPath)
}

func Test_parseLibraryConfig(t *testing.T) {
	cases := []struct {
		name     string
		expected LibraryConfig
		isErr    bool
	}{
		{
			name:     "parts-infra/contour",
			expected: LibraryConfig{Registry: "parts-infra", Name: "contour"},
		},
		{
			name:     "contour",
			expected: LibraryConfig{Name: "contour"},
		},
		{
			name:     "parts-infra/contour@0.1.0",
			expected: LibraryConfig{Registry: "parts-infra", Name: "contour", Version: "0.1.0"},
		},
		{
			name:     "contour@0.1.0",
			expected: LibraryConfig{Registry: "", Name: "contour", Version: "0.1.0"},
		},
		{
			name:  "@foo/bar@baz@doh",
			isErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			d, err := parseLibraryConfig(tc.name)
			if tc.isErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.expected, d)
		})
	}
}
