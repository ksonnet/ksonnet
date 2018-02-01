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
	"path"
	"strings"

	param "github.com/ksonnet/ksonnet/metadata/params"
	"github.com/ksonnet/ksonnet/prototype"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

func (m *manager) ComponentPaths() (AbsPaths, error) {
	paths := AbsPaths{}
	err := afero.Walk(m.appFS, string(m.componentsPath), func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Only add file paths and exclude the params.libsonnet file
		if !info.IsDir() && path.Base(p) != componentParamsFile {
			paths = append(paths, p)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return paths, nil
}

func (m *manager) GetAllComponents() ([]string, error) {
	componentPaths, err := m.ComponentPaths()
	if err != nil {
		return nil, err
	}

	var components []string
	for _, p := range componentPaths {
		component := strings.TrimSuffix(path.Base(p), path.Ext(p))
		components = append(components, component)
	}

	return components, nil
}

func (m *manager) CreateComponent(name string, text string, params param.Params, templateType prototype.TemplateType) error {
	if !isValidName(name) || strings.Contains(name, "/") {
		return fmt.Errorf("Component name '%s' is not valid; must not contain punctuation, spaces, or begin or end with a slash", name)
	}

	componentPath := string(appendToAbsPath(m.componentsPath, name))
	switch templateType {
	case prototype.YAML:
		componentPath = componentPath + ".yaml"
	case prototype.JSON:
		componentPath = componentPath + ".json"
	case prototype.Jsonnet:
		componentPath = componentPath + ".jsonnet"
	default:
		return fmt.Errorf("Unrecognized prototype template type '%s'", templateType)
	}

	if exists, err := afero.Exists(m.appFS, componentPath); exists {
		return fmt.Errorf("Component with name '%s' already exists", name)
	} else if err != nil {
		return fmt.Errorf("Could not check whether component '%s' exists:\n\n%v", name, err)
	}

	log.Infof("Writing component at '%s/%s'", componentsDir, name)
	err := afero.WriteFile(m.appFS, componentPath, []byte(text), defaultFilePermissions)
	if err != nil {
		return err
	}

	log.Debugf("Writing component parameters at '%s/%s", componentsDir, name)
	return m.writeComponentParams(name, params)
}

// DeleteComponent removes the component file and all references.
// Write operations will happen at the end to minimalize failures that leave
// the directory structure in a half-finished state.
func (m *manager) DeleteComponent(name string) error {
	componentPath, err := m.findComponentPath(name)
	if err != nil {
		return err
	}

	// Build the new component/params.libsonnet file.
	componentParamsFile, err := afero.ReadFile(m.appFS, string(m.componentParamsPath))
	if err != nil {
		return err
	}
	componentJsonnet, err := param.DeleteComponent(name, string(componentParamsFile))
	if err != nil {
		return err
	}

	// Build the new environment/<env>/params.libsonnet files.
	// environment name -> jsonnet
	envJsonnets := make(map[string]string)
	envs, err := m.GetEnvironments()
	if err != nil {
		return err
	}
	for _, env := range envs {
		path := appendToAbsPath(m.environmentsPath, env.Name, paramsFileName)
		envParamsFile, err := afero.ReadFile(m.appFS, string(path))
		if err != nil {
			return err
		}
		jsonnet, err := param.DeleteEnvironmentComponent(name, string(envParamsFile))
		if err != nil {
			return err
		}
		envJsonnets[env.Name] = jsonnet
	}

	//
	// Delete the component references.
	//
	log.Infof("Removing component parameter references ...")

	// Remove the references in component/params.libsonnet.
	log.Debugf("... deleting references in %s", m.componentParamsPath)
	err = afero.WriteFile(m.appFS, string(m.componentParamsPath), []byte(componentJsonnet), defaultFilePermissions)
	if err != nil {
		return err
	}
	// Remove the component references in each environment's
	// environment/<env>/params.libsonnet.
	for _, env := range envs {
		path := appendToAbsPath(m.environmentsPath, env.Name, paramsFileName)
		log.Debugf("... deleting references in %s", path)
		err = afero.WriteFile(m.appFS, string(path), []byte(envJsonnets[env.Name]), defaultFilePermissions)
		if err != nil {
			return err
		}
	}

	//
	// Delete the component file in components/.
	//
	log.Infof("Deleting component '%s' at path '%s'", name, componentPath)
	if err := m.appFS.Remove(componentPath); err != nil {
		return err
	}

	// TODO: Remove,
	// references in main.jsonnet.
	// component references in other component files (feature does not yet exist).
	log.Infof("Succesfully deleted component '%s'", name)
	return nil
}

func (m *manager) GetComponentParams(component string) (param.Params, error) {
	text, err := afero.ReadFile(m.appFS, string(m.componentParamsPath))
	if err != nil {
		return nil, err
	}

	return param.GetComponentParams(component, string(text))
}

func (m *manager) GetAllComponentParams() (map[string]param.Params, error) {
	text, err := afero.ReadFile(m.appFS, string(m.componentParamsPath))
	if err != nil {
		return nil, err
	}

	return param.GetAllComponentParams(string(text))
}

func (m *manager) SetComponentParams(component string, params param.Params) error {
	text, err := afero.ReadFile(m.appFS, string(m.componentParamsPath))
	if err != nil {
		return err
	}

	jsonnet, err := param.SetComponentParams(component, string(text), params)
	if err != nil {
		return err
	}

	return afero.WriteFile(m.appFS, string(m.componentParamsPath), []byte(jsonnet), defaultFilePermissions)
}

func (m *manager) findComponentPath(name string) (string, error) {
	componentPaths, err := m.ComponentPaths()
	if err != nil {
		log.Debugf("Failed to retrieve component paths")
		return "", err
	}

	var componentPath string
	for _, p := range componentPaths {
		fileName := path.Base(p)
		component := strings.TrimSuffix(fileName, path.Ext(fileName))

		if component == name {
			// need to make sure we don't have multiple files with the same component name
			if componentPath != "" {
				return "", fmt.Errorf("Found multiple component files with component name '%s'", name)
			}
			componentPath = p
		}
	}

	if componentPath == "" {
		return "", fmt.Errorf("No component with name '%s' found", name)
	}

	return componentPath, nil
}

func (m *manager) writeComponentParams(componentName string, params param.Params) error {
	text, err := afero.ReadFile(m.appFS, string(m.componentParamsPath))
	if err != nil {
		return err
	}

	appended, err := param.AppendComponent(componentName, string(text), params)
	if err != nil {
		return err
	}

	return afero.WriteFile(m.appFS, string(m.componentParamsPath), []byte(appended), defaultFilePermissions)
}

func genComponentParamsContent() []byte {
	return []byte(`{
  global: {
    // User-defined global parameters; accessible to all component and environments, Ex:
    // replicas: 4,
  },
  components: {
    // Component-level parameters, defined initially from 'ks prototype use ...'
    // Each object below should correspond to a component in the components/ directory
  },
}
`)
}
