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
package lib

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/spf13/afero"

	str "github.com/ksonnet/ksonnet/strings"
)

const (
	blankSwagger     = "/blankSwagger.json"
	blankSwaggerData = `{
  "swagger": "2.0",
  "info": {
   "title": "Kubernetes",
   "version": "v1.7.0"
  },
  "paths": {
  },
  "definitions": {
  }
}`
)

var testFS = afero.NewMemMapFs()

func init() {
	afero.WriteFile(testFS, blankSwagger, []byte(blankSwaggerData), os.ModePerm)
}

func TestGenerateLibData(t *testing.T) {
	specFlag := fmt.Sprintf("file:%s", blankSwagger)
	libPath := "lib"

	libManager, err := NewManager(specFlag, testFS, libPath)
	if err != nil {
		t.Fatal("Failed to initialize lib.Manager")
	}

	err = libManager.GenerateLibData()
	if err != nil {
		t.Fatal("Failed to generate lib data")
	}

	// Verify contents of lib.
	versionPath := str.AppendToPath(libPath, "v1.7.0")

	schemaPath := str.AppendToPath(versionPath, schemaFilename)
	bytes, err := afero.ReadFile(testFS, string(schemaPath))
	if err != nil {
		t.Fatalf("Failed to read swagger file at '%s':\n%v", schemaPath, err)
	}

	if actualSwagger := string(bytes); actualSwagger != blankSwaggerData {
		t.Fatalf("Expected swagger file at '%s' to have value: '%s', got: '%s'", schemaPath, blankSwaggerData, actualSwagger)
	}

	k8sLibPath := str.AppendToPath(versionPath, k8sLibFilename)
	k8sLibBytes, err := afero.ReadFile(testFS, string(k8sLibPath))
	if err != nil {
		t.Fatalf("Failed to read ksonnet-lib file at '%s':\n%v", k8sLibPath, err)
	}

	blankK8sLib, err := ioutil.ReadFile("testdata/k8s.libsonnet")
	if err != nil {
		t.Fatalf("Failed to read testdata/k8s.libsonnet: %#v", err)
	}

	if actualK8sLib := string(k8sLibBytes); actualK8sLib != string(blankK8sLib) {
		t.Fatalf("Expected swagger file at '%s' to have value: '%s', got: '%s'", k8sLibPath, blankK8sLib, actualK8sLib)
	}
}
