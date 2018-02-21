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

package component

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	param "github.com/ksonnet/ksonnet/metadata/params"
	"github.com/ksonnet/ksonnet/prototype"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

var (
	// defaultFolderPermissions are the default permissions for a folder.
	defaultFolderPermissions = os.FileMode(0755)
	// defaultFilePermissions are the default permission for a file.
	defaultFilePermissions = os.FileMode(0644)
)

// Create creates a component.
func Create(fs afero.Fs, root, name, text string, params param.Params, templateType prototype.TemplateType) (string, error) {
	cc, err := newComponentCreator(fs, root)
	if err != nil {
		return "", errors.Wrap(err, "initialize component creator")
	}

	return cc.Create(name, text, params, templateType)
}

type componentCreator struct {
	fs   afero.Fs
	root string
}

func newComponentCreator(fs afero.Fs, root string) (*componentCreator, error) {
	if fs == nil {
		return nil, errors.New("fs is nil")
	}

	if root == "" {
		return nil, errors.New("invalid ksonnet root")
	}

	return &componentCreator{fs: fs, root: root}, nil
}

func (cc *componentCreator) Create(name, text string, params param.Params, templateType prototype.TemplateType) (string, error) {
	if !isValidName(name) {
		return "", errors.Errorf("Component name '%s' is not valid; must not contain punctuation, spaces, or begin or end with a slash", name)
	}

	nsName, componentName := namespaceComponent(name)

	componentDir, componentPath, err := cc.location(nsName, componentName, templateType)
	if err != nil {
		return "", errors.Wrap(err, "generate component location")
	}

	paramsPath := filepath.Join(componentDir, "params.libsonnet")

	exists, err := afero.Exists(cc.fs, componentDir)
	if err != nil {
		return "", errors.Wrapf(err, "check if %s exists", componentDir)
	}

	if !exists {
		if err = cc.createNamespace(componentDir, paramsPath); err != nil {
			return "", err
		}
	}

	exists, err = afero.Exists(cc.fs, componentPath)
	if err != nil {
		return "", errors.Wrapf(err, "check if %s exists", componentPath)
	}

	if exists {
		return "", errors.Errorf("component with name '%s' already exists", name)
	}

	log.Infof("Writing component at '%s'", componentPath)
	if err := afero.WriteFile(cc.fs, componentPath, []byte(text), defaultFilePermissions); err != nil {
		return "", errors.Wrapf(err, "write component %s")
	}

	log.Debugf("Writing component parameters at '%s/%s", componentsRoot, name)

	if err := cc.writeParams(componentName, paramsPath, params); err != nil {
		return "", errors.Wrapf(err, "write parameters")
	}

	return componentPath, nil
}

// location returns the dir and full path for the component.
func (cc *componentCreator) location(nsName, name string, templateType prototype.TemplateType) (string, string, error) {
	componentDir := filepath.Join(cc.root, componentsRoot, nsName)
	componentPath := filepath.Join(componentDir, name)
	switch templateType {
	case prototype.YAML:
		componentPath = componentPath + ".yaml"
	case prototype.JSON:
		componentPath = componentPath + ".json"
	case prototype.Jsonnet:
		componentPath = componentPath + ".jsonnet"
	default:
		return "", "", errors.Errorf("Unrecognized prototype template type '%s'", templateType)
	}

	return componentDir, componentPath, nil
}

func (cc *componentCreator) createNamespace(componentDir, paramsPath string) error {
	if err := cc.fs.MkdirAll(componentDir, defaultFolderPermissions); err != nil {
		return errors.Wrapf(err, "create component dir %s", componentDir)
	}

	if err := afero.WriteFile(cc.fs, paramsPath, GenParamsContent(), defaultFilePermissions); err != nil {
		return errors.Wrap(err, "create component params")
	}

	return nil
}

func (cc *componentCreator) writeParams(name, paramsPath string, params param.Params) error {
	text, err := afero.ReadFile(cc.fs, paramsPath)
	if err != nil {
		return err
	}

	appended, err := param.AppendComponent(name, string(text), params)
	if err != nil {
		return err
	}

	return afero.WriteFile(cc.fs, paramsPath, []byte(appended), defaultFilePermissions)
}

// isValidName returns true if a name (e.g., for an environment) is valid.
// A component is valid if it does not contain punctuation, whitespace, leading or
// trailing slashes.
func isValidName(name string) bool {
	// No unicode whitespace is allowed. `Fields` doesn't handle trailing or
	// leading whitespace.
	fields := strings.Fields(name)
	if len(fields) > 1 || len(strings.TrimSpace(name)) != len(name) {
		return false
	}

	hasPunctuation := regexp.MustCompile(`[\\,;':!()?"{}\[\]*&%@$]+`).MatchString
	hasTrailingSlashes := regexp.MustCompile(`/+$`).MatchString
	hasLeadingSlashes := regexp.MustCompile(`^/+`).MatchString
	return len(name) != 0 && !hasPunctuation(name) && !hasTrailingSlashes(name) && !hasLeadingSlashes(name)
}

func namespaceComponent(name string) (string, string) {
	parts := strings.Split(name, "/")

	if len(parts) == 1 {
		return "", parts[0]
	}

	var nsName []string
	var componentName string
	for i := range parts {
		if i == len(parts)-1 {
			componentName = parts[i]
			break
		}

		nsName = append(nsName, parts[i])
	}

	return strings.Join(nsName, "/"), componentName
}

// GenParamsContent is the default content for params.libsonnet.
func GenParamsContent() []byte {
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
