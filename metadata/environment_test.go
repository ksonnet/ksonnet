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

package metadata

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"

	param "github.com/ksonnet/ksonnet/metadata/params"
	"github.com/spf13/afero"
)

const (
	mockSpecJSONServer = "localhost:8080"

	mockEnvName  = "us-west/test"
	mockEnvName2 = "us-west/prod"
	mockEnvName3 = "us-east/test"
)

var mockAPIServer = "http://example.com"
var mockNamespace = "some-namespace"
var mockEnvs = []string{defaultEnvName, mockEnvName, mockEnvName2, mockEnvName3}

func mockEnvironments(t *testing.T, appName string) *manager {
	return mockEnvironmentsWith(t, appName, mockEnvs)
}

func mockEnvironmentsWith(t *testing.T, appName string, envNames []string) *manager {
	spec, err := parseClusterSpec(fmt.Sprintf("file:%s", blankSwagger), testFS)
	if err != nil {
		t.Fatalf("Failed to parse cluster spec: %v", err)
	}

	appPath := AbsPath(appName)
	reg := newMockRegistryManager("incubator")
	m, err := initManager(appName, appPath, spec, &mockAPIServer, &mockNamespace, reg, testFS)
	if err != nil {
		t.Fatalf("Failed to init cluster spec: %v", err)
	}

	for _, env := range envNames {
		envPath := appendToAbsPath(m.environmentsPath, env)
		testFS.Mkdir(string(envPath), defaultFolderPermissions)
		testDirExists(t, string(envPath))

		envFilePath := appendToAbsPath(envPath, envFileName)
		envFileData := m.generateOverrideData()
		err = afero.WriteFile(testFS, string(envFilePath), envFileData, defaultFilePermissions)
		if err != nil {
			t.Fatalf("Could not write file at path: %s", envFilePath)
		}
		testFileExists(t, string(envFilePath))

		specPath := appendToAbsPath(envPath, specFilename)
		specData, err := generateSpecData(mockSpecJSONServer, mockNamespace)
		if err != nil {
			t.Fatalf("Expected to marshal:\nserver: %s\nnamespace: %s\n, but failed", mockSpecJSONServer, mockNamespace)
		}
		err = afero.WriteFile(testFS, string(specPath), specData, defaultFilePermissions)
		if err != nil {
			t.Fatalf("Could not write file at path: %s", specPath)
		}
		testFileExists(t, string(specPath))

		paramsPath := appendToAbsPath(envPath, paramsFileName)
		paramsData := m.generateParamsData()
		err = afero.WriteFile(testFS, string(paramsPath), paramsData, defaultFilePermissions)
		if err != nil {
			t.Fatalf("Could not write file at path: %s", paramsPath)
		}
		testFileExists(t, string(paramsPath))
	}

	return m
}

func testDirExists(t *testing.T, path string) {
	exists, err := afero.DirExists(testFS, path)
	if err != nil {
		t.Fatalf("Expected directory at '%s' to exist, but failed:\n%v", path, err)
	} else if !exists {
		t.Fatalf("Expected directory at '%s' to exist, but it does not", path)
	}
}

func testDirNotExists(t *testing.T, path string) {
	exists, err := afero.DirExists(testFS, path)
	if err != nil {
		t.Fatalf("Expected directory at '%s' to be removed, but failed:\n%v", path, err)
	} else if exists {
		t.Fatalf("Expected directory at '%s' to be removed, but it exists", path)
	}
}

func testFileExists(t *testing.T, path string) {
	exists, err := afero.Exists(testFS, path)
	if err != nil {
		t.Fatalf("Expected file at '%s' to exist, but failed:\n%v", path, err)
	} else if !exists {
		t.Fatalf("Expected file at '%s' to exist, but it does not", path)
	}
}

func TestDeleteEnvironment(t *testing.T) {
	appName := "test-delete-envs"
	m := mockEnvironments(t, appName)

	// Test that both directory and empty parent directory is deleted.
	expectedPath := appendToAbsPath(m.environmentsPath, mockEnvName3)
	parentDir := strings.Split(mockEnvName3, "/")[0]
	expectedParentPath := appendToAbsPath(m.environmentsPath, parentDir)
	err := m.DeleteEnvironment(mockEnvName3)
	if err != nil {
		t.Fatalf("Expected %s to be deleted but got err:\n  %s", mockEnvName3, err)
	}
	testDirNotExists(t, string(expectedPath))
	testDirNotExists(t, string(expectedParentPath))

	// Test that only leaf directory is deleted if parent directory is shared
	expectedPath = appendToAbsPath(m.environmentsPath, mockEnvName2)
	parentDir = strings.Split(mockEnvName2, "/")[0]
	expectedParentPath = appendToAbsPath(m.environmentsPath, parentDir)
	err = m.DeleteEnvironment(mockEnvName2)
	if err != nil {
		t.Fatalf("Expected %s to be deleted but got err:\n  %s", mockEnvName3, err)
	}
	testDirNotExists(t, string(expectedPath))
	testDirExists(t, string(expectedParentPath))
}

func TestGetEnvironments(t *testing.T) {
	m := mockEnvironments(t, "test-get-envs")

	envs, err := m.GetEnvironments()
	if err != nil {
		t.Fatalf("Expected to successfully get environments but failed:\n  %s", err)
	}

	if len(envs) != 4 {
		t.Fatalf("Expected to get %d environments, got %d", 4, len(envs))
	}

	if envs[0].Server != mockSpecJSONServer {
		t.Fatalf("Expected env server to be %s, got %s", mockSpecJSONServer, envs[0].Server)
	}
}

func TestSetEnvironment(t *testing.T) {
	appName := "test-set-envs"
	m := mockEnvironments(t, appName)

	setName := "new-env"
	setServer := "http://example.com"
	setNamespace := "some-namespace"
	set := Environment{Name: setName, Server: setServer, Namespace: setNamespace}

	// Test updating an environment that doesn't exist
	err := m.SetEnvironment("notexists", &set)
	if err == nil {
		t.Fatal("Expected error when setting an environment that does not exist")
	}

	// Test updating an environment to an environment that already exists
	err = m.SetEnvironment(mockEnvName, &Environment{Name: mockEnvName2})
	if err == nil {
		t.Fatalf("Expected error when setting \"%s\" to \"%s\", because env already exists", mockEnvName, mockEnvName2)
	}

	//
	// Test changing the name and server of a an existing environment.
	//

	err = m.SetEnvironment(mockEnvName, &set)
	if err != nil {
		t.Fatalf("Could not set \"%s\", got:\n  %s", mockEnvName, err)
	}

	// Ensure new env directory is created, and old directory no longer exists.
	envPath := appendToAbsPath(AbsPath(appName), environmentsDir)
	expectedPathExists := appendToAbsPath(envPath, set.Name)
	expectedPathNotExists := appendToAbsPath(envPath, mockEnvName)
	testDirExists(t, string(expectedPathExists))
	testDirNotExists(t, string(expectedPathNotExists))

	// BUG: https://github.com/spf13/afero/issues/141
	// we aren't able to test this until the above is fixed.
	//
	// ensure all files are moved
	//
	// expectedFiles := []string{
	// 	envFileName,
	// 	specFilename,
	// 	paramsFileName,
	// }
	// for _, f := range expectedFiles {
	// 	expectedFilePath := appendToAbsPath(expectedPathExists, f)
	// 	testFileExists(t, string(expectedFilePath))
	// }

	// ensure spec file contains the correct content
	specData, err := afero.ReadFile(testFS, string(appendToAbsPath(expectedPathExists, specFilename)))
	if err != nil {
		t.Fatalf("Failed to read spec file:\n  %s", err)
	}
	var envSpec EnvironmentSpec
	err = json.Unmarshal(specData, &envSpec)
	if err != nil {
		t.Fatalf("Failed to read spec file:\n  %s", err)
	}
	if envSpec.Server != set.Server {
		t.Fatalf("Expected server to be set to '%s', got: '%s'", set.Server, envSpec.Server)
	}
	if envSpec.Namespace != set.Namespace {
		t.Fatalf("Expected namespace to be set to '%s', got: '%s'", set.Namespace, envSpec.Namespace)
	}

	tests := []struct {
		appName string
		nameOld string
		nameNew string
	}{
		// Test changing the name of an env 'us-west' to 'us-west/dev'
		{
			"test-set-to-child",
			"us-west",
			"us-west/dev",
		},
		// Test changing the name of an env 'us-west/dev' to 'us-west'
		{
			"test-set-to-parent",
			"us-west/dev",
			"us-west",
		},
	}

	for _, v := range tests {
		m = mockEnvironmentsWith(t, v.appName, []string{v.nameOld})
		err = m.SetEnvironment(v.nameOld, &Environment{Name: v.nameNew})
		if err != nil {
			t.Fatalf("Could not set '%s', got:\n  %s", v.nameOld, err)
		}
		// Ensure new env directory is created
		expectedPath := appendToAbsPath(AbsPath(v.appName), environmentsDir, v.nameNew)
		testDirExists(t, string(expectedPath))
	}
}

func TestGenerateOverrideData(t *testing.T) {
	m := mockEnvironments(t, "test-gen-override-data")

	expected := `local base = import "../base.libsonnet";
local k = import "k.libsonnet";

base + {
  // Insert user-specified overrides here. For example if a component is named "nginx-deployment", you might have something like:
  //   "nginx-deployment"+: k.deployment.mixin.metadata.labels({foo: "bar"})
}
`
	result := m.generateOverrideData()

	if string(result) != expected {
		t.Fatalf("Expected to generate override file with data:\n%s\n,got:\n%s", expected, result)
	}
}

func TestGenerateParamsData(t *testing.T) {
	m := mockEnvironments(t, "test-gen-params-data")

	expected := `local params = import "../../components/params.libsonnet";
params + {
  components +: {
    // Insert component parameter overrides here. Ex:
    // guestbook +: {
    //   name: "guestbook-dev",
    //   replicas: params.global.replicas,
    // },
  },
}
`
	result := string(m.generateParamsData())

	if result != expected {
		t.Fatalf("Expected to generate params file with data:\n%s\n, got:\n%s", expected, result)
	}
}

func TestMergeParamMaps(t *testing.T) {
	tests := []struct {
		base      map[string]param.Params
		overrides map[string]param.Params
		expected  map[string]param.Params
	}{
		{
			map[string]param.Params{
				"bar": param.Params{"replicas": "5"},
			},
			map[string]param.Params{
				"foo": param.Params{"name": `"foo"`, "replicas": "1"},
			},
			map[string]param.Params{
				"bar": param.Params{"replicas": "5"},
				"foo": param.Params{"name": `"foo"`, "replicas": "1"},
			},
		},
		{
			map[string]param.Params{
				"bar": param.Params{"replicas": "5"},
			},
			map[string]param.Params{
				"bar": param.Params{"name": `"foo"`},
			},
			map[string]param.Params{
				"bar": param.Params{"name": `"foo"`, "replicas": "5"},
			},
		},
		{
			map[string]param.Params{
				"bar": param.Params{"name": `"bar"`, "replicas": "5"},
				"foo": param.Params{"name": `"foo"`, "replicas": "4"},
				"baz": param.Params{"name": `"baz"`, "replicas": "3"},
			},
			map[string]param.Params{
				"foo": param.Params{"replicas": "1"},
				"baz": param.Params{"name": `"foobaz"`},
			},
			map[string]param.Params{
				"bar": param.Params{"name": `"bar"`, "replicas": "5"},
				"foo": param.Params{"name": `"foo"`, "replicas": "1"},
				"baz": param.Params{"name": `"foobaz"`, "replicas": "3"},
			},
		},
	}

	for _, s := range tests {
		result := mergeParamMaps(s.base, s.overrides)
		if !reflect.DeepEqual(s.expected, result) {
			t.Errorf("Wrong merge\n  expected:\n%v\n  got:\n%v", s.expected, result)
		}
	}
}
