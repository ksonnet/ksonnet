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
	"strings"

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
	exists, err := m.environmentExists(name)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("Environment '%s' already exists", name)
	}

	envPath := appendToAbsPath(m.environmentsDir, name)
	err = m.appFS.MkdirAll(string(envPath), os.ModePerm)
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

func (m *manager) GetEnvironments() ([]Environment, error) {
	envs := []Environment{}

	err := afero.Walk(m.appFS, string(m.environmentsDir), func(path string, f os.FileInfo, err error) error {
		isDir, err := afero.IsDir(m.appFS, path)
		if err != nil {
			return err
		}

		if isDir {
			// Only want leaf directories containing a spec.json
			specPath := filepath.Join(path, specFilename)
			specFileExists, err := afero.Exists(m.appFS, specPath)
			if err != nil {
				return err
			}
			if specFileExists {
				envName := filepath.Clean(strings.TrimPrefix(path, string(m.environmentsDir)+"/"))
				specFile, err := afero.ReadFile(m.appFS, specPath)
				if err != nil {
					return err
				}
				var envSpec EnvironmentSpec
				err = json.Unmarshal(specFile, &envSpec)
				if err != nil {
					return err
				}

				envs = append(envs, Environment{Name: envName, Path: path, URI: envSpec.URI})
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return envs, nil
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
