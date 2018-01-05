// Copyright 2017 The ksonnet authors
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
	"sort"
	"testing"
)

func TestComponentPaths(t *testing.T) {
	spec, err := parseClusterSpec(fmt.Sprintf("file:%s", blankSwagger), testFS)
	if err != nil {
		t.Fatalf("Failed to parse cluster spec: %v", err)
	}

	appPath := AbsPath("/componentPaths")
	reg := newMockRegistryManager("incubator")
	m, err := initManager("componentPaths", appPath, spec, &mockAPIServer, &mockNamespace, reg, testFS)
	if err != nil {
		t.Fatalf("Failed to init cluster spec: %v", err)
	}

	// Create empty app file.
	components := appendToAbsPath(appPath, componentsDir)
	appFile1 := appendToAbsPath(components, "component1.jsonnet")
	f1, err := testFS.OpenFile(string(appFile1), os.O_RDONLY|os.O_CREATE, 0777)
	if err != nil {
		t.Fatalf("Failed to touch app file '%s'\n%v", appFile1, err)
	}
	f1.Close()

	// Create empty file in a nested directory.
	appSubdir := appendToAbsPath(components, "appSubdir")
	err = testFS.MkdirAll(string(appSubdir), os.ModePerm)
	if err != nil {
		t.Fatalf("Failed to create directory '%s'\n%v", appSubdir, err)
	}
	appFile2 := appendToAbsPath(appSubdir, "component2.jsonnet")
	f2, err := testFS.OpenFile(string(appFile2), os.O_RDONLY|os.O_CREATE, 0777)
	if err != nil {
		t.Fatalf("Failed to touch app file '%s'\n%v", appFile1, err)
	}
	f2.Close()

	// Create a directory that won't be listed in the call to `ComponentPaths`.
	unlistedDir := string(appendToAbsPath(components, "doNotListMe"))
	err = testFS.MkdirAll(unlistedDir, os.ModePerm)
	if err != nil {
		t.Fatalf("Failed to create directory '%s'\n%v", unlistedDir, err)
	}

	paths, err := m.ComponentPaths()
	if err != nil {
		t.Fatalf("Failed to find component paths: %v", err)
	}

	sort.Slice(paths, func(i, j int) bool { return paths[i] < paths[j] })

	if len(paths) != 3 || paths[0] != string(appFile2) || paths[1] != string(appFile1) {
		t.Fatalf("m.ComponentPaths failed; expected '%s', got '%s'", []string{string(appFile1), string(appFile2)}, paths)
	}
}
