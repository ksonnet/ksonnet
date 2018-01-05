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
