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
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"

	"github.com/ksonnet/ksonnet-lib/ksonnet-gen/ksonnet"
	"github.com/ksonnet/ksonnet-lib/ksonnet-gen/kubespec"
	param "github.com/ksonnet/ksonnet/metadata/params"
)

const (
	defaultEnvName  = "default"
	metadataDirName = ".metadata"

	// hidden metadata files
	schemaFilename        = "swagger.json"
	extensionsLibFilename = "k.libsonnet"
	k8sLibFilename        = "k8s.libsonnet"

	// primary environment files
	envFileName    = "main.jsonnet"
	paramsFileName = "params.libsonnet"
	specFilename   = "spec.json"
)

var envPaths = []string{
	// metadata Dir.wh
	metadataDirName,
	// environment base override file
	envFileName,
	// params file
	paramsFileName,
	// spec file
	specFilename,
}

// Environment represents all fields of a ksonnet environment
type Environment struct {
	Path      string
	Name      string
	Server    string
	Namespace string
}

// EnvironmentSpec represents the contents in spec.json.
type EnvironmentSpec struct {
	Server    string `json:"server"`
	Namespace string `json:"namespace"`
}

func (m *manager) CreateEnvironment(name, server, namespace string, spec ClusterSpec) error {
	extensionsLibData, k8sLibData, specData, err := m.generateKsonnetLibData(spec)
	if err != nil {
		log.Debugf("Failed to write '%s'", specFilename)
		return err
	}

	err = m.createEnvironment(name, server, namespace, extensionsLibData, k8sLibData, specData)
	if err == nil {
		log.Infof("Environment '%s' pointing to namespace '%s' and server address at '%s' successfully created", name, namespace, server)
	}
	return err
}

func (m *manager) createEnvironment(name, server, namespace string, extensionsLibData, k8sLibData, specData []byte) error {
	exists, err := m.environmentExists(name)
	if err != nil {
		log.Debug("Failed to check whether environment exists")
		return err
	}
	if exists {
		return fmt.Errorf("Environment '%s' already exists", name)
	}

	// ensure environment name does not contain punctuation
	if !isValidName(name) {
		return fmt.Errorf("Environment name '%s' is not valid; must not contain punctuation, spaces, or begin or end with a slash", name)
	}

	log.Infof("Creating environment '%s' with namespace '%s', pointing at server at address '%s'", name, namespace, server)

	envPath := appendToAbsPath(m.environmentsPath, name)
	err = m.appFS.MkdirAll(string(envPath), defaultFolderPermissions)
	if err != nil {
		return err
	}

	metadataPath := appendToAbsPath(envPath, metadataDirName)
	err = m.appFS.MkdirAll(string(metadataPath), defaultFolderPermissions)
	if err != nil {
		return err
	}

	log.Infof("Generating environment metadata at path '%s'", envPath)

	// Generate the environment spec file.
	envSpecData, err := generateSpecData(server, namespace)
	if err != nil {
		return err
	}

	metadata := []struct {
		path AbsPath
		data []byte
	}{
		{
			// schema file
			appendToAbsPath(metadataPath, schemaFilename),
			specData,
		},
		{
			// k8s file
			appendToAbsPath(metadataPath, k8sLibFilename),
			k8sLibData,
		},
		{
			// extensions file
			appendToAbsPath(metadataPath, extensionsLibFilename),
			extensionsLibData,
		},
		{
			// environment base override file
			appendToAbsPath(envPath, envFileName),
			m.generateOverrideData(),
		},
		{
			// params file
			appendToAbsPath(envPath, paramsFileName),
			m.generateParamsData(),
		},
		{
			// spec file
			appendToAbsPath(envPath, specFilename),
			envSpecData,
		},
	}

	for _, a := range metadata {
		fileName := path.Base(string(a.path))
		log.Debugf("Generating '%s', length: %d", fileName, len(a.data))
		if err = afero.WriteFile(m.appFS, string(a.path), a.data, defaultFilePermissions); err != nil {
			log.Debugf("Failed to write '%s'", fileName)
			return err
		}
	}

	return nil
}

func (m *manager) DeleteEnvironment(name string) error {
	envPath := string(appendToAbsPath(m.environmentsPath, name))

	// Check whether this environment exists
	envExists, err := m.environmentExists(name)
	if err != nil {
		log.Debug("Failed to check whether environment exists")
		return err
	}
	if !envExists {
		return fmt.Errorf("Environment '%s' does not exist", name)
	}

	log.Infof("Deleting environment '%s' at path '%s'", name, envPath)

	// Remove the directory and all files within the environment path.
	err = m.appFS.RemoveAll(envPath)
	if err != nil {
		log.Debugf("Failed to remove environment directory at path '%s'", envPath)
		return err
	}

	// Need to ensure empty parent directories are also removed.
	log.Debug("Removing empty parent directories, if any")
	err = m.cleanEmptyParentDirs(name)
	if err != nil {
		return err
	}

	log.Infof("Successfully removed environment '%s'", name)
	return nil
}

func (m *manager) GetEnvironments() ([]*Environment, error) {
	envs := []*Environment{}

	log.Debug("Retrieving all environments")
	err := afero.Walk(m.appFS, string(m.environmentsPath), func(path string, f os.FileInfo, err error) error {
		isDir, err := afero.IsDir(m.appFS, path)
		if err != nil {
			log.Debugf("Failed to check whether the path at '%s' is a directory", path)
			return err
		}

		if isDir {
			// Only want leaf directories containing a spec.json
			specPath := filepath.Join(path, specFilename)
			specFileExists, err := afero.Exists(m.appFS, specPath)
			if err != nil {
				log.Debugf("Failed to check whether spec file at '$s' exists", specPath)
				return err
			}
			if specFileExists {
				envName := filepath.Clean(strings.TrimPrefix(path, string(m.environmentsPath)+"/"))
				specFile, err := afero.ReadFile(m.appFS, specPath)
				if err != nil {
					log.Debugf("Failed to read spec file at path '%s'", specPath)
					return err
				}
				var envSpec EnvironmentSpec
				err = json.Unmarshal(specFile, &envSpec)
				if err != nil {
					log.Debugf("Failed to convert the spec file at path '%s' to JSON", specPath)
					return err
				}

				log.Debugf("Found environment '%s', with server '%s' and namespace '%s'", envName, envSpec.Server, envSpec.Namespace)
				envs = append(envs, &Environment{Name: envName, Path: path, Server: envSpec.Server, Namespace: envSpec.Namespace})
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return envs, nil
}

func (m *manager) GetEnvironment(name string) (*Environment, error) {
	envs, err := m.GetEnvironments()
	if err != nil {
		return nil, err
	}

	for _, env := range envs {
		if env.Name == name {
			return env, nil
		}
	}

	return nil, fmt.Errorf("Environment '%s' does not exist", name)
}

func (m *manager) SetEnvironment(name string, desired *Environment) error {
	env, err := m.GetEnvironment(name)
	if err != nil {
		return err
	}

	// If the name has changed, the directory location needs to be moved to
	// reflect the change.
	if name != desired.Name && len(desired.Name) != 0 {
		// ensure new environment name does not contain punctuation
		if !isValidName(desired.Name) {
			return fmt.Errorf("Environment name '%s' is not valid; must not contain punctuation, spaces, or begin or end with a slash", name)
		}

		log.Infof("Setting environment name from '%s' to '%s'", name, desired.Name)

		// Ensure not overwriting another environment
		desiredExists, err := m.environmentExists(desired.Name)
		if err != nil {
			log.Debugf("Failed to check whether environment '%s' already exists", desired.Name)
			return err
		}
		if desiredExists {
			return fmt.Errorf("Can not update '%s' to '%s', it already exists", name, desired.Name)
		}

		//
		// Move the directory
		//

		pathOld := appendToAbsPath(m.environmentsPath, name)
		pathNew := appendToAbsPath(m.environmentsPath, desired.Name)
		exists, err := afero.DirExists(m.appFS, string(pathNew))
		if err != nil {
			return err
		}

		if exists {
			// we know that the desired path is not an environment from
			// the check earlier. This is an intermediate directory.
			// We need to move the file contents.
			m.tryMvEnvDir(pathOld, pathNew)
		} else if filepath.HasPrefix(string(pathNew), string(pathOld)) {
			// the new directory is a child of the old directory --
			// rename won't work.
			err = m.appFS.MkdirAll(string(pathNew), defaultFolderPermissions)
			if err != nil {
				return err
			}
			m.tryMvEnvDir(pathOld, pathNew)
		} else {
			// Need to first create subdirectories that don't exist
			intermediatePath := path.Dir(string(pathNew))
			log.Debugf("Moving directory at path '%s' to '%s'", string(pathOld), string(pathNew))
			err = m.appFS.MkdirAll(intermediatePath, defaultFolderPermissions)
			if err != nil {
				return err
			}
			// finally, move the directory
			err = m.appFS.Rename(string(pathOld), string(pathNew))
			if err != nil {
				log.Debugf("Failed to move path '%s' to '%s", string(pathOld), string(pathNew))
				return err
			}
		}

		// clean up any empty parent directory paths
		err = m.cleanEmptyParentDirs(name)
		if err != nil {
			return err
		}
		name = desired.Name
	}

	//
	// Update fields in spec.json.
	//

	var server string
	if len(desired.Server) != 0 {
		log.Infof("Setting environment server to '%s'", desired.Server)
		server = desired.Server
	} else {
		server = env.Server
	}
	var namespace string
	if len(desired.Namespace) != 0 {
		log.Infof("Setting environment namespace to '%s'", desired.Namespace)
		namespace = desired.Namespace
	} else {
		namespace = env.Namespace
	}

	newSpec, err := generateSpecData(server, namespace)
	if err != nil {
		log.Debugf("Failed to generate %s with server '%s' and namespace '%s'", specFilename, server, namespace)
		return err
	}

	envPath := appendToAbsPath(m.environmentsPath, name)
	specPath := appendToAbsPath(envPath, specFilename)

	err = afero.WriteFile(m.appFS, string(specPath), newSpec, defaultFilePermissions)
	if err != nil {
		log.Debugf("Failed to write %s at path '%s'", specFilename, specPath)
		return err
	}

	log.Infof("Successfully updated environment '%s'", name)
	return nil
}

func (m *manager) GetEnvironmentParams(name string) (map[string]param.Params, error) {
	exists, err := m.environmentExists(name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("Environment '%s' does not exist", name)
	}

	// Get the environment specific params
	envParamsPath := appendToAbsPath(m.environmentsPath, name, paramsFileName)
	envParamsText, err := afero.ReadFile(m.appFS, string(envParamsPath))
	if err != nil {
		return nil, err
	}
	envParams, err := param.GetAllEnvironmentParams(string(envParamsText))
	if err != nil {
		return nil, err
	}

	// Get all component params
	componentParams, err := m.GetAllComponentParams()
	if err != nil {
		return nil, err
	}

	// Merge the param sets, replacing the component params if the environment params override
	return mergeParamMaps(componentParams, envParams), nil
}

func (m *manager) SetEnvironmentParams(env, component string, params param.Params) error {
	exists, err := m.environmentExists(env)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("Environment '%s' does not exist", env)
	}

	path := appendToAbsPath(m.environmentsPath, env, paramsFileName)

	text, err := afero.ReadFile(m.appFS, string(path))
	if err != nil {
		return err
	}

	appended, err := param.SetEnvironmentParams(component, string(text), params)
	if err != nil {
		return err
	}

	err = afero.WriteFile(m.appFS, string(path), []byte(appended), defaultFilePermissions)
	if err != nil {
		return err
	}

	log.Debugf("Successfully set parameters for component '%s' at environment '%s'", component, env)
	return nil
}

func (m *manager) tryMvEnvDir(dirPathOld, dirPathNew AbsPath) error {
	// first ensure none of these paths exists in the new directory
	for _, p := range envPaths {
		path := string(appendToAbsPath(dirPathNew, p))
		if exists, err := afero.Exists(m.appFS, path); err != nil {
			return err
		} else if exists {
			return fmt.Errorf("%s already exists", path)
		}
	}

	// note: afero and go does not provide simple ways to move the
	// contents. We'll have to rename them individually.
	for _, p := range envPaths {
		err := m.appFS.Rename(string(appendToAbsPath(dirPathOld, p)), string(appendToAbsPath(dirPathNew, p)))
		if err != nil {
			return err
		}
	}
	// clean up the old directory if it is empty
	if empty, err := afero.IsEmpty(m.appFS, string(dirPathOld)); err != nil {
		return err
	} else if empty {
		return m.appFS.RemoveAll(string(dirPathOld))
	}
	return nil
}

func (m *manager) cleanEmptyParentDirs(name string) error {
	// clean up any empty parent directory paths
	log.Debug("Removing empty parent directories, if any")
	parentDir := name
	for parentDir != "." {
		parentDir = filepath.Dir(parentDir)
		parentPath := string(appendToAbsPath(m.environmentsPath, parentDir))

		isEmpty, err := afero.IsEmpty(m.appFS, parentPath)
		if err != nil {
			log.Debugf("Failed to check whether parent directory at path '%s' is empty", parentPath)
			return err
		}
		if isEmpty {
			log.Debugf("Failed to remove parent directory at path '%s'", parentPath)
			err := m.appFS.RemoveAll(parentPath)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *manager) generateKsonnetLibData(spec ClusterSpec) ([]byte, []byte, []byte, error) {
	// Get cluster specification data, possibly from the network.
	text, err := spec.data()
	if err != nil {
		return nil, nil, nil, err
	}

	ksonnetLibDir := appendToAbsPath(m.environmentsPath, defaultEnvName)

	// Deserialize the API object.
	s := kubespec.APISpec{}
	err = json.Unmarshal(text, &s)
	if err != nil {
		return nil, nil, nil, err
	}

	s.Text = text
	s.FilePath = filepath.Dir(string(ksonnetLibDir))

	// Emit Jsonnet code.
	extensionsLibData, k8sLibData, err := ksonnet.Emit(&s, nil, nil)
	return extensionsLibData, k8sLibData, text, err
}

func (m *manager) generateOverrideData() []byte {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("local base = import \"%s\";\n", m.baseLibsonnetPath))
	buf.WriteString(fmt.Sprintf("local k = import \"%s\";\n\n", extensionsLibFilename))
	buf.WriteString("base + {\n")
	buf.WriteString("  // Insert user-specified overrides here. For example if a component is named \"nginx-deployment\", you might have something like:\n")
	buf.WriteString("  //   \"nginx-deployment\"+: k.deployment.mixin.metadata.labels({foo: \"bar\"})\n")
	buf.WriteString("}\n")
	return buf.Bytes()
}

func (m *manager) generateParamsData() []byte {
	return []byte(`local params = import "` + m.componentParamsPath + `";
params + {
  components +: {
    // Insert component parameter overrides here. Ex:
    // guestbook +: {
    //   name: "guestbook-dev",
    //   replicas: params.global.replicas,
    // },
  },
}
`)
}

func generateSpecData(server, namespace string) ([]byte, error) {
	// Format the spec json and return; preface keys with 2 space idents.
	return json.MarshalIndent(EnvironmentSpec{Server: server, Namespace: namespace}, "", "  ")
}

func (m *manager) environmentExists(name string) (bool, error) {
	envs, err := m.GetEnvironments()
	if err != nil {
		return false, err
	}

	envExists := false
	for _, env := range envs {
		if env.Name == name {
			envExists = true
			break
		}
	}

	return envExists, nil
}

func mergeParamMaps(base, overrides map[string]param.Params) map[string]param.Params {
	for component, params := range overrides {
		if _, contains := base[component]; !contains {
			base[component] = params
		} else {
			for k, v := range params {
				base[component][k] = v
			}
		}
	}
	return base
}
