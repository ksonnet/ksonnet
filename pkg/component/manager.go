// Copyright 2018 The ksonnet authors
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
	"path/filepath"
	"strings"

	"github.com/ksonnet/ksonnet/pkg/app"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
)

// ResolvePath resolves a given path to a module and a component.
func ResolvePath(ksApp app.App, path string) (Module, Component, error) {
	isDir, err := isComponentDir2(ksApp, path)
	if err != nil {
		return nil, nil, errors.Wrap(err, "checking for module directory")
	}

	if isDir {
		m, err := GetModule(ksApp, path)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "retrieving module %s", path)
		}

		return m, nil, nil
	}

	moduleName, componentName, err := extractPathParts(ksApp, path)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "extracting module and component names from %s", path)
	}

	m, err := GetModule(ksApp, moduleName)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "retrieving module %s", path)
	}

	c, err := LocateComponent(ksApp, moduleName, componentName)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "locating component %q in module %q", componentName, moduleName)
	}

	return m, c, nil
}

var (
	// DefaultManager is the default manager for components.
	DefaultManager = &defaultManager{}
)

// Manager is an interface for interacting with components.
type Manager interface {
	Components(ns Module) ([]Component, error)
	Component(ksApp app.App, module, componentName string) (Component, error)
	CreateModule(ksApp app.App, name string) error
	Module(ksApp app.App, moduleName string) (Module, error)
	Modules(ksApp app.App, envName string) ([]Module, error)
}

type defaultManager struct{}

var _ Manager = (*defaultManager)(nil)

func (dm *defaultManager) Modules(ksApp app.App, envName string) ([]Module, error) {
	return ModulesFromEnv(ksApp, envName)
}

func (dm *defaultManager) Module(ksApp app.App, module string) (Module, error) {
	return GetModule(ksApp, module)
}

func (dm *defaultManager) Components(ns Module) ([]Component, error) {
	return ns.Components()
}

func (dm *defaultManager) Component(ksApp app.App, module, componentName string) (Component, error) {
	return LocateComponent(ksApp, module, componentName)
}

func (dm *defaultManager) CreateModule(ksApp app.App, name string) error {
	parts := strings.Split(name, ".")
	dir := filepath.Join(append([]string{ksApp.Root(), "components"}, parts...)...)

	if err := ksApp.Fs().MkdirAll(dir, app.DefaultFolderPermissions); err != nil {
		return err
	}

	paramsDir := filepath.Join(dir, "params.libsonnet")
	return afero.WriteFile(ksApp.Fs(), paramsDir, GenParamsContent(), app.DefaultFilePermissions)
}

func isComponentDir2(ksApp app.App, path string) (bool, error) {
	parts := strings.Split(path, "/")
	dir := filepath.Join(append([]string{ksApp.Root(), componentsRoot}, parts...)...)
	dir = filepath.Clean(dir)

	return afero.DirExists(ksApp.Fs(), dir)
}

// extractPathParts extracts the module and component name from a path.
func extractPathParts(ksApp app.App, path string) (string, string, error) {
	if strings.Contains(path, "/") {
		return "", "", errors.New("component can't contain a /")
	}

	path = strings.Replace(path, ".", string(filepath.Separator), -1)
	module, componentName := extractModuleComponent(ksApp, path)
	base := filepath.Join(module.Dir(), componentName)

	exts := []string{".yaml", ".jsonnet", ".json"}
	for _, ext := range exts {
		exists, err := afero.Exists(ksApp.Fs(), base+ext)
		if err != nil {
			return "", "", errors.Wrap(err, "check for component")
		}

		if exists {
			return module.Name(), componentName, nil
		}
	}

	return "", "", errors.Errorf("%q is not a component or a module", path)
}
