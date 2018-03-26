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
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ksonnet/ksonnet/metadata/app"
	"github.com/ksonnet/ksonnet/metadata/app/mocks"
	str "github.com/ksonnet/ksonnet/strings"
	"github.com/stretchr/testify/require"

	"github.com/spf13/afero"
)

const (
	mockEnvName  = "us-west/test"
	mockEnvName2 = "us-west/prod"
	mockEnvName3 = "us-east/test"
)

var (
	mockAPIServer = "http://example.com"
	mockNamespace = "some-namespace"
	mockEnvs      = []string{defaultEnvName, mockEnvName, mockEnvName2, mockEnvName3}
)

func mockEnvironments(t *testing.T, fs afero.Fs, appName string) *manager {
	return mockEnvironmentsWith(t, fs, appName, mockEnvs)
}

func mockEnvironmentsWith(t *testing.T, fs afero.Fs, appName string, envNames []string) *manager {
	specFlag := fmt.Sprintf("file:%s", blankSwagger)

	reg := newMockRegistryManager("incubator")
	root := filepath.Join("/", appName)
	m, err := initManager(appName, root, &specFlag, &mockAPIServer, &mockNamespace, reg, fs)
	if err != nil {
		t.Fatalf("Failed to init cluster spec: %v", err)
	}

	for _, env := range envNames {
		envPath := str.AppendToPath(m.environmentsPath, env)
		fs.Mkdir(envPath, defaultFolderPermissions)
		testDirExists(t, fs, envPath)

		envFilePath := str.AppendToPath(envPath, envFileName)
		envFileData := m.generateOverrideData()
		err = afero.WriteFile(fs, envFilePath, envFileData, defaultFilePermissions)
		if err != nil {
			t.Fatalf("Could not write file at path: %s", envFilePath)
		}
		testFileExists(t, fs, envFilePath)

		paramsPath := str.AppendToPath(envPath, paramsFileName)
		paramsData := m.generateParamsData()
		err = afero.WriteFile(fs, paramsPath, paramsData, defaultFilePermissions)
		if err != nil {
			t.Fatalf("Could not write file at path: %s", paramsPath)
		}
		testFileExists(t, fs, paramsPath)

		appSpec, err := app.Read(m.appFS, m.rootPath)
		if err != nil {
			t.Fatal("Could not retrieve app spec")
		}
		appSpec.AddEnvironmentSpec(&app.EnvironmentSpec{
			Name:              env,
			Path:              env,
			KubernetesVersion: "v1.8.7",
			Destination: &app.EnvironmentDestinationSpec{
				Server:    mockAPIServer,
				Namespace: mockNamespace,
			},
		})
		err = app.Write(m.appFS, m.rootPath, appSpec)
		require.NoError(t, err)
	}

	return m
}

func testDirExists(t *testing.T, fs afero.Fs, path string) {
	exists, err := afero.DirExists(fs, path)
	require.NoError(t, err, "Checking %q failed", path)
	require.True(t, exists, "Expected directory %q to exist", path)
}

func testDirNotExists(t *testing.T, fs afero.Fs, path string) {
	exists, err := afero.DirExists(fs, path)
	require.NoError(t, err, "Checking %q failed", path)
	require.False(t, exists, "Expected directory %q to not exist", path)
}

func testFileExists(t *testing.T, fs afero.Fs, path string) {
	exists, err := afero.Exists(fs, path)
	require.NoError(t, err, "Checking %q failed", path)
	require.True(t, exists, "Expected file %q to exist", path)
}

func TestDeleteEnvironment(t *testing.T) {
	withFs(func(fs afero.Fs) {
		appName := "test-delete-envs"
		appMock := &mocks.App{}
		appMock.On("RemoveEnvironment", "us-east/test").Return(nil)
		appMock.On("RemoveEnvironment", "us-west/prod").Return(nil)

		m := mockEnvironments(t, fs, appName)

		// Test that both directory and empty parent directory is deleted.
		expectedPath, err := filepath.Abs(filepath.Join("/", m.rootPath, "environments", mockEnvName3))
		require.NoError(t, err)
		parentDir := strings.Split(mockEnvName3, "/")[0]
		expectedParentPath, err := filepath.Abs(filepath.Join("/", m.rootPath, "environments", parentDir))
		require.NoError(t, err)
		err = m.DeleteEnvironment(mockEnvName3)
		if err != nil {
			t.Fatalf("Expected %s to be deleted but got err:\n  %s", mockEnvName3, err)
		}
		testDirNotExists(t, fs, expectedPath)
		testDirNotExists(t, fs, expectedParentPath)

		// Test that only leaf directory is deleted if parent directory is shared
		expectedPath = str.AppendToPath("/", m.environmentsPath, mockEnvName2)
		parentDir = strings.Split(mockEnvName2, "/")[0]
		expectedParentPath = str.AppendToPath("/", m.environmentsPath, parentDir)
		err = m.DeleteEnvironment(mockEnvName2)
		if err != nil {
			t.Fatalf("Expected %s to be deleted but got err:\n  %s", mockEnvName3, err)
		}

		testDirNotExists(t, fs, expectedPath)
		testDirExists(t, fs, expectedParentPath)
	})
}

func TestGetEnvironments(t *testing.T) {
	withFs(func(fs afero.Fs) {
		appMock := &mocks.App{}
		appMock.On("Environments").Return(nil, nil)

		m := mockEnvironments(t, fs, "test-get-envs")

		envs, err := m.GetEnvironments()
		if err != nil {
			t.Fatalf("Expected to successfully get environments but failed:\n  %s", err)
		}

		if len(envs) != 4 {
			t.Fatalf("Expected to get %d environments, got %d", 4, len(envs))
		}

		cur := envs[mockEnvName]
		name := cur.Name
		if name != mockEnvName {
			t.Fatalf("Expected env name to be %q, got %q", mockEnvName, name)
		}

		server := cur.Destination.Server()
		if server != mockAPIServer {
			t.Fatalf("Expected env server to be %q, got %q", mockAPIServer, server)
		}
	})
}

func TestSetEnvironment(t *testing.T) {
	withFs(func(fs afero.Fs) {
		appName := "test-set-envs"
		m := mockEnvironments(t, fs, appName)

		setName := "new-env"

		desired := Environment{
			Name: mockEnvName2,
		}

		// Test updating an environment to an environment that already exists
		err := m.SetEnvironment(mockEnvName, desired)
		if err == nil {
			t.Fatalf("Expected error when setting \"%s\" to \"%s\", because env already exists", mockEnvName, mockEnvName2)
		}

		desired = Environment{
			Name: setName,
		}

		// Test changing the name an existing environment.
		err = m.SetEnvironment(mockEnvName, desired)
		if err != nil {
			t.Fatalf("Could not set \"%s\", got:\n  %s", mockEnvName, err)
		}

		// Ensure new env directory is created, and old directory no longer exists.
		envPath := str.AppendToPath(appName, environmentsDir)
		expectedPathExists := filepath.Join("/", envPath, setName)
		expectedPathNotExists := filepath.Join("/", envPath, mockEnvName)
		testDirExists(t, fs, expectedPathExists)
		testDirNotExists(t, fs, expectedPathNotExists)

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
			m = mockEnvironmentsWith(t, fs, v.appName, []string{v.nameOld})
			desired := Environment{Name: v.nameNew}
			err = m.SetEnvironment(v.nameOld, desired)
			if err != nil {
				t.Fatalf("Could not set '%s', got:\n  %s", v.nameOld, err)
			}
			// Ensure new env directory is created
			expectedPath := filepath.Join("/", v.appName, environmentsDir, v.nameNew)
			testDirExists(t, fs, expectedPath)
		}
	})
}

func TestGenerateOverrideData(t *testing.T) {
	withFs(func(fs afero.Fs) {
		m := mockEnvironments(t, fs, "test-gen-override-data")

		expected := `local base = import "base.libsonnet";
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
	})
}

func TestGenerateParamsData(t *testing.T) {
	withFs(func(fs afero.Fs) {
		m := mockEnvironments(t, fs, "test-gen-params-data")

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
	})
}
