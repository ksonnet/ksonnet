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
	"fmt"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"strings"

	"github.com/ksonnet/ksonnet/metadata/app"
	param "github.com/ksonnet/ksonnet/metadata/params"
	"github.com/ksonnet/ksonnet/prototype"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

func appendToAbsPath(originalPath AbsPath, toAppend ...string) AbsPath {
	paths := append([]string{string(originalPath)}, toAppend...)
	return AbsPath(path.Join(paths...))
}

const (
	ksonnetDir      = ".ksonnet"
	libDir          = "lib"
	componentsDir   = "components"
	environmentsDir = "environments"
	vendorDir       = "vendor"

	componentParamsFile = "params.libsonnet"
	baseLibsonnetFile   = "base.libsonnet"
	appYAMLFile         = "app.yaml"

	// ComponentsExtCodeKey is the ExtCode key for component imports
	ComponentsExtCodeKey = "__ksonnet/components"
	// ParamsExtCodeKey is the ExtCode key for importing environment parameters
	ParamsExtCodeKey = "__ksonnet/params"

	// User-level ksonnet directories.
	userKsonnetRootDir = ".ksonnet"
	pkgSrcCacheDir     = "src"
)

type manager struct {
	appFS afero.Fs

	// Application paths.
	rootPath         AbsPath
	ksonnetPath      AbsPath
	libPath          AbsPath
	componentsPath   AbsPath
	environmentsPath AbsPath
	vendorPath       AbsPath

	componentParamsPath AbsPath
	baseLibsonnetPath   AbsPath
	appYAMLPath         AbsPath

	// User-level paths.
	userKsonnetRootPath AbsPath
	pkgSrcCachePath     AbsPath
}

func findManager(abs AbsPath, appFS afero.Fs) (*manager, error) {
	var lastBase string
	currBase := string(abs)

	for {
		currPath := path.Join(currBase, ksonnetDir)
		exists, err := afero.Exists(appFS, currPath)
		if err != nil {
			return nil, err
		}
		if exists {
			return newManager(AbsPath(currBase), appFS)
		}

		lastBase = currBase
		currBase = filepath.Dir(currBase)
		if lastBase == currBase {
			return nil, fmt.Errorf("No ksonnet application found")
		}
	}
}

func initManager(name string, rootPath AbsPath, spec ClusterSpec, serverURI, namespace *string, appFS afero.Fs) (*manager, error) {
	m, err := newManager(rootPath, appFS)
	if err != nil {
		return nil, err
	}

	// Generate the program text for ksonnet-lib.
	//
	// IMPLEMENTATION NOTE: We get the cluster specification and generate
	// ksonnet-lib before initializing the directory structure so that failure of
	// either (e.g., GET'ing the spec from a live cluster returns 404) does not
	// result in a partially-initialized directory structure.
	//
	extensionsLibData, k8sLibData, specData, err := m.generateKsonnetLibData(spec)
	if err != nil {
		return nil, err
	}

	// Initialize directory structure.
	if err := m.createAppDirTree(name); err != nil {
		return nil, err
	}

	// Initialize user dir structure.
	if err := m.createUserDirTree(); err != nil {
		return nil, err
	}

	// Initialize environment, and cache specification data.
	if serverURI != nil {
		err := m.createEnvironment(defaultEnvName, *serverURI, *namespace, extensionsLibData, k8sLibData, specData)
		if err != nil {
			return nil, err
		}
	}

	return m, nil
}

func newManager(rootPath AbsPath, appFS afero.Fs) (*manager, error) {
	usr, err := user.Current()
	if err != nil {
		return nil, err
	}
	userRootPath := appendToAbsPath(AbsPath(usr.HomeDir), userKsonnetRootDir)

	return &manager{
		appFS: appFS,

		// Application paths.
		rootPath:         rootPath,
		ksonnetPath:      appendToAbsPath(rootPath, ksonnetDir),
		libPath:          appendToAbsPath(rootPath, libDir),
		componentsPath:   appendToAbsPath(rootPath, componentsDir),
		environmentsPath: appendToAbsPath(rootPath, environmentsDir),
		vendorPath:       appendToAbsPath(rootPath, vendorDir),

		componentParamsPath: appendToAbsPath(rootPath, componentsDir, componentParamsFile),
		baseLibsonnetPath:   appendToAbsPath(rootPath, environmentsDir, baseLibsonnetFile),
		appYAMLPath:         appendToAbsPath(rootPath, appYAMLFile),

		// User-level paths.
		userKsonnetRootPath: userRootPath,
		pkgSrcCachePath:     appendToAbsPath(userRootPath, pkgSrcCacheDir),
	}, nil
}

func (m *manager) Root() AbsPath {
	return m.rootPath
}

func (m *manager) ComponentPaths() (AbsPaths, error) {
	paths := AbsPaths{}
	err := afero.Walk(m.appFS, string(m.componentsPath), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			paths = append(paths, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return paths, nil
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

func (m *manager) LibPaths(envName string) (libPath, envLibPath, envComponentPath, envParamsPath AbsPath) {
	envPath := appendToAbsPath(m.environmentsPath, envName)
	return m.libPath, appendToAbsPath(envPath, metadataDirName),
		appendToAbsPath(envPath, path.Base(envName)+".jsonnet"), appendToAbsPath(envPath, componentParamsFile)
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

func (m *manager) createUserDirTree() error {
	dirPaths := []AbsPath{
		m.userKsonnetRootPath,
		m.pkgSrcCachePath,
	}

	for _, p := range dirPaths {
		if err := m.appFS.MkdirAll(string(p), defaultFolderPermissions); err != nil {
			return err
		}
	}

	return nil
}

func (m *manager) createAppDirTree(name string) error {
	exists, err := afero.DirExists(m.appFS, string(m.rootPath))
	if err != nil {
		return fmt.Errorf("Could not check existance of directory '%s':\n%v", m.rootPath, err)
	} else if exists {
		return fmt.Errorf("Could not create app; directory '%s' already exists", m.rootPath)
	}

	dirPaths := []AbsPath{
		m.rootPath,
		m.ksonnetPath,
		m.libPath,
		m.componentsPath,
		m.environmentsPath,
		m.vendorPath,
	}

	for _, p := range dirPaths {
		if err := m.appFS.MkdirAll(string(p), defaultFolderPermissions); err != nil {
			return err
		}
	}

	appYAML, err := genAppYAMLContent(name)
	if err != nil {
		return err
	}

	filePaths := []struct {
		path    AbsPath
		content []byte
	}{
		{
			m.componentParamsPath,
			genComponentParamsContent(),
		},
		{
			m.baseLibsonnetPath,
			genBaseLibsonnetContent(),
		},
		{
			m.appYAMLPath,
			appYAML,
		},
	}

	for _, f := range filePaths {
		if err := afero.WriteFile(m.appFS, string(f.path), f.content, defaultFilePermissions); err != nil {
			return err
		}
	}

	return nil
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

func genAppYAMLContent(name string) ([]byte, error) {
	content := app.Spec{
		APIVersion: app.DefaultAPIVersion,
		Kind:       app.Kind,
		Name:       name,
		Version:    app.DefaultVersion,
	}

	return json.MarshalIndent(&content, "", "  ")
}

func genBaseLibsonnetContent() []byte {
	return []byte(`local components = std.extVar("` + ComponentsExtCodeKey + `");
components + {
  // Insert user-specified overrides here.
}
`)
}
