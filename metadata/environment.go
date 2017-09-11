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
	"os"
	"path/filepath"

	"github.com/spf13/afero"

	"github.com/ksonnet/ksonnet-lib/ksonnet-gen/ksonnet"
	"github.com/ksonnet/ksonnet-lib/ksonnet-gen/kubespec"
)

const (
	defaultEnvName = "default"

	schemaFilename        = "swagger.json"
	extensionsLibFilename = "k.libsonnet"
	k8sLibFilename        = "k8s.libsonnet"
	specFilename          = "spec.json"
)

type Environment struct {
	Path string
	Name string
	URI  string
}

type EnvironmentSpec struct {
	URI string `json:"uri"`
}

func (m *manager) CreateEnvironment(name, uri string, spec ClusterSpec, extensionsLibData, k8sLibData []byte) error {
	envPath := appendToAbsPath(m.environmentsDir, name)
	err := m.appFS.MkdirAll(string(envPath), os.ModePerm)
	if err != nil {
		return err
	}

	// Get cluster specification data, possibly from the network.
	specData, err := spec.data()
	if err != nil {
		return err
	}

	// Generate the schema file.
	schemaPath := appendToAbsPath(envPath, schemaFilename)
	err = afero.WriteFile(m.appFS, string(schemaPath), specData, os.ModePerm)
	if err != nil {
		return err
	}

	k8sLibPath := appendToAbsPath(envPath, k8sLibFilename)
	err = afero.WriteFile(m.appFS, string(k8sLibPath), k8sLibData, 0644)
	if err != nil {
		return err
	}

	extensionsLibPath := appendToAbsPath(envPath, extensionsLibFilename)
	err = afero.WriteFile(m.appFS, string(extensionsLibPath), extensionsLibData, 0644)
	if err != nil {
		return err
	}

	// Generate the environment spec file.
	envSpecData, err := generateSpecData(uri)
	if err != nil {
		return err
	}

	envSpecPath := appendToAbsPath(envPath, specFilename)
	return afero.WriteFile(m.appFS, string(envSpecPath), envSpecData, os.ModePerm)
}

func (m *manager) GenerateKsonnetLibData(spec ClusterSpec) ([]byte, []byte, error) {
	// Get cluster specification data, possibly from the network.
	text, err := spec.data()
	if err != nil {
		return nil, nil, err
	}

	ksonnetLibDir := appendToAbsPath(m.environmentsDir, defaultEnvName)

	// Deserialize the API object.
	s := kubespec.APISpec{}
	err = json.Unmarshal(text, &s)
	if err != nil {
		return nil, nil, err
	}

	s.Text = text
	s.FilePath = filepath.Dir(string(ksonnetLibDir))

	// Emit Jsonnet code.
	return ksonnet.Emit(&s, nil, nil)
}

func generateSpecData(uri string) ([]byte, error) {
	// Format the spec json and return; preface keys with 2 space idents.
	return json.MarshalIndent(EnvironmentSpec{URI: uri}, "", "  ")
}
