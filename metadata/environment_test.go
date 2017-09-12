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
	"os"
	"testing"

	"github.com/spf13/afero"
)

const (
	mockSpecJSON    = "spec.json"
	mockSpecJSONURI = "localhost:8080"
	mockEnvName     = "us-west/test"
)

func TestGetEnvironments(t *testing.T) {
	spec, err := parseClusterSpec(fmt.Sprintf("file:%s", blankSwagger), testFS)
	if err != nil {
		t.Fatalf("Failed to parse cluster spec: %v", err)
	}

	appPath := AbsPath("/test-app")
	m, err := initManager(appPath, spec, testFS)
	if err != nil {
		t.Fatalf("Failed to init cluster spec: %v", err)
	}

	defaultEnvDir := appendToAbsPath(environmentsDir, defaultEnvName)
	anotherEnvDir := appendToAbsPath(environmentsDir, mockEnvName)

	path := appendToAbsPath(appPath, environmentsDir)
	exists, err := afero.DirExists(testFS, string(path))
	if err != nil {
		t.Fatalf("Expected to create directory '%s', but failed:\n%v", environmentsDir, err)
	} else if !exists {
		t.Fatalf("Expected to create directory '%s', but failed", path)
	}

	defaultEnvPath := appendToAbsPath(appPath, string(defaultEnvDir))
	anotherEnvPath := appendToAbsPath(appPath, string(anotherEnvDir))
	specDefaultEnvPath := appendToAbsPath(defaultEnvPath, mockSpecJSON)
	specAnotherEnvPath := appendToAbsPath(anotherEnvPath, mockSpecJSON)

	specData, err := generateSpecData(mockSpecJSONURI)
	if err != nil {
		t.Fatalf("Expected to marshal:\n%s\n, but failed", mockSpecJSONURI)
	}

	err = afero.WriteFile(testFS, string(specDefaultEnvPath), specData, os.ModePerm)
	if err != nil {
		t.Fatalf("Could not write file at path: %s", specDefaultEnvPath)
	}
	err = afero.WriteFile(testFS, string(specAnotherEnvPath), specData, os.ModePerm)
	if err != nil {
		t.Fatalf("Could not write file at path: %s", specAnotherEnvPath)
	}

	envs, err := m.GetEnvironments()
	if err != nil {
		t.Fatalf("Expected to successfully get environments but failed:\n  %s", err)
	}

	if len(envs) != 2 {
		t.Fatalf("Expected to get %d environments, got %d", 2, len(envs))
	}

	if envs[0].URI != mockSpecJSONURI {
		t.Fatalf("Expected env URI to be %s, got %s", mockSpecJSONURI, envs[0].URI)
	}
}
