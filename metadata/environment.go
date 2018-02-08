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
	"fmt"
	"path"
	"path/filepath"

	"github.com/ksonnet/ksonnet/metadata/app"
	"github.com/ksonnet/ksonnet/metadata/lib"
	str "github.com/ksonnet/ksonnet/strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"

	param "github.com/ksonnet/ksonnet/metadata/params"
)

const (
	defaultEnvName = "default"

	// primary environment files
	envFileName    = "main.jsonnet"
	paramsFileName = "params.libsonnet"
)

var envPaths = []string{
	// environment base override file
	envFileName,
	// params file
	paramsFileName,
}

func (m *manager) CreateEnvironment(name, server, namespace, k8sSpecFlag string) error {
	// generate the lib data for this kubernetes version
	libManager, err := lib.NewManagerWithSpec(k8sSpecFlag, m.appFS, m.libPath)
	if err != nil {
		return err
	}

	if err := libManager.GenerateLibData(); err != nil {
		return err
	}

	// add the environment to the app spec
	appSpec, err := m.AppSpec()
	if err != nil {
		return err
	}

	if _, exists := appSpec.GetEnvironmentSpec(name); exists {
		return fmt.Errorf("Environment '%s' already exists", name)
	}

	// ensure environment name does not contain punctuation
	if !isValidName(name) {
		return fmt.Errorf("Environment name '%s' is not valid; must not contain punctuation, spaces, or begin or end with a slash", name)
	}

	if namespace == "" {
		namespace = "default"
	}

	log.Infof("Creating environment '%s' with namespace '%s', pointing at server at address '%s'", name, namespace, server)

	envPath := str.AppendToPath(m.environmentsPath, name)
	err = m.appFS.MkdirAll(envPath, defaultFolderPermissions)
	if err != nil {
		return err
	}

	metadata := []struct {
		path string
		data []byte
	}{
		{
			// environment base override file
			str.AppendToPath(envPath, envFileName),
			m.generateOverrideData(),
		},
		{
			// params file
			str.AppendToPath(envPath, paramsFileName),
			m.generateParamsData(),
		},
	}

	for _, a := range metadata {
		fileName := path.Base(a.path)
		log.Debugf("Generating '%s', length: %d", fileName, len(a.data))
		if err = afero.WriteFile(m.appFS, a.path, a.data, defaultFilePermissions); err != nil {
			log.Debugf("Failed to write '%s'", fileName)
			return err
		}
	}

	// update app.yaml
	err = appSpec.AddEnvironmentSpec(&app.EnvironmentSpec{
		Name: name,
		Path: name,
		Destinations: app.EnvironmentDestinationSpecs{
			&app.EnvironmentDestinationSpec{
				Server:    server,
				Namespace: namespace,
			},
		},
		KubernetesVersion: libManager.K8sVersion,
	})

	if err != nil {
		return err
	}

	return m.WriteAppSpec(appSpec)
}

func (m *manager) DeleteEnvironment(name string) error {
	app, err := m.AppSpec()
	if err != nil {
		return err
	}

	env, ok := app.GetEnvironmentSpec(name)
	if !ok {
		return fmt.Errorf("Environment '%s' does not exist", name)
	}

	envPath := str.AppendToPath(m.environmentsPath, env.Path)

	log.Infof("Deleting environment '%s' with metadata at path '%s'", name, envPath)

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

	// Update app spec.
	if err := m.WriteAppSpec(app); err != nil {
		return err
	}

	log.Infof("Successfully removed environment '%s'", name)
	return nil
}

func (m *manager) GetEnvironments() (app.EnvironmentSpecs, error) {
	app, err := m.AppSpec()
	if err != nil {
		return nil, err
	}

	log.Debug("Retrieving all environments")
	return app.GetEnvironmentSpecs(), nil
}

func (m *manager) GetEnvironment(name string) (*app.EnvironmentSpec, error) {
	app, err := m.AppSpec()
	if err != nil {
		return nil, err
	}

	env, ok := app.GetEnvironmentSpec(name)
	if !ok {
		return nil, fmt.Errorf("Environment '%s' does not exist", name)
	}

	return env, nil
}

func (m *manager) SetEnvironment(name, desiredName string) error {
	if name == desiredName || len(desiredName) == 0 {
		return nil
	}

	// ensure new environment name does not contain punctuation
	if !isValidName(desiredName) {
		return fmt.Errorf("Environment name '%s' is not valid; must not contain punctuation, spaces, or begin or end with a slash", name)
	}

	// Ensure not overwriting another environment
	desiredExists, err := m.environmentExists(desiredName)
	if err != nil {
		log.Debugf("Failed to check whether environment '%s' already exists", desiredName)
		return err
	}
	if desiredExists {
		return fmt.Errorf("Failed to update '%s'; environment '%s' exists", name, desiredName)
	}

	log.Infof("Setting environment name from '%s' to '%s'", name, desiredName)

	//
	// Update app spec. We will write out the app spec changes once all file
	// move operations are complete.
	//

	appSpec, err := m.AppSpec()
	if err != nil {
		return err
	}

	current, exists := appSpec.GetEnvironmentSpec(name)
	if !exists {
		return fmt.Errorf("Trying to update an environment that doesn't exist")
	}

	err = appSpec.UpdateEnvironmentSpec(name, &app.EnvironmentSpec{
		Name:              desiredName,
		Destinations:      current.Destinations,
		KubernetesVersion: current.KubernetesVersion,
		Targets:           current.Targets,
		Path:              desiredName,
	})

	if err != nil {
		return err
	}

	//
	// If the name has changed, the directory location needs to be moved to
	// reflect the change.
	//

	pathOld := str.AppendToPath(m.environmentsPath, name)
	pathNew := str.AppendToPath(m.environmentsPath, desiredName)
	exists, err = afero.DirExists(m.appFS, pathNew)
	if err != nil {
		return err
	}

	if exists {
		// we know that the desired path is not an environment from
		// the check earlier. This is an intermediate directory.
		// We need to move the file contents.
		m.tryMvEnvDir(pathOld, pathNew)
	} else if filepath.HasPrefix(pathNew, pathOld) {
		// the new directory is a child of the old directory --
		// rename won't work.
		err = m.appFS.MkdirAll(pathNew, defaultFolderPermissions)
		if err != nil {
			return err
		}
		m.tryMvEnvDir(pathOld, pathNew)
	} else {
		// Need to first create subdirectories that don't exist
		intermediatePath := path.Dir(pathNew)
		log.Debugf("Moving directory at path '%s' to '%s'", pathOld, pathNew)
		err = m.appFS.MkdirAll(intermediatePath, defaultFolderPermissions)
		if err != nil {
			return err
		}
		// finally, move the directory
		err = m.appFS.Rename(pathOld, pathNew)
		if err != nil {
			log.Debugf("Failed to move path '%s' to '%s", pathOld, pathNew)
			return err
		}
	}

	// clean up any empty parent directory paths
	err = m.cleanEmptyParentDirs(name)
	if err != nil {
		return err
	}

	m.WriteAppSpec(appSpec)

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
	envParamsPath := str.AppendToPath(m.environmentsPath, name, paramsFileName)
	envParamsText, err := afero.ReadFile(m.appFS, envParamsPath)
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

	path := str.AppendToPath(m.environmentsPath, env, paramsFileName)

	text, err := afero.ReadFile(m.appFS, path)
	if err != nil {
		return err
	}

	appended, err := param.SetEnvironmentParams(component, string(text), params)
	if err != nil {
		return err
	}

	err = afero.WriteFile(m.appFS, path, []byte(appended), defaultFilePermissions)
	if err != nil {
		return err
	}

	log.Debugf("Successfully set parameters for component '%s' at environment '%s'", component, env)
	return nil
}

func (m *manager) EnvPaths(env string) (libPath, mainPath, paramsPath string, err error) {
	app, err := m.AppSpec()
	if err != nil {
		return
	}

	envSpec, ok := app.GetEnvironmentSpec(env)
	if !ok {
		err = fmt.Errorf("Environment '%s' does not exist", env)
		return
	}

	libManager := lib.NewManager(envSpec.KubernetesVersion, m.appFS, m.libPath)

	envPath := str.AppendToPath(m.environmentsPath, env)

	// main.jsonnet file
	mainPath = str.AppendToPath(envPath, envFileName)
	// params.libsonnet file
	paramsPath = str.AppendToPath(envPath, componentParamsFile)
	// ksonnet-lib file directory
	libPath, err = libManager.GetLibPath()
	return
}

func (m *manager) tryMvEnvDir(dirPathOld, dirPathNew string) error {
	// first ensure none of these paths exists in the new directory
	for _, p := range envPaths {
		path := str.AppendToPath(dirPathNew, p)
		if exists, err := afero.Exists(m.appFS, path); err != nil {
			return err
		} else if exists {
			return fmt.Errorf("%s already exists", path)
		}
	}

	// note: afero and go does not provide simple ways to move the
	// contents. We'll have to rename them individually.
	for _, p := range envPaths {
		err := m.appFS.Rename(str.AppendToPath(dirPathOld, p), str.AppendToPath(dirPathNew, p))
		if err != nil {
			return err
		}
	}
	// clean up the old directory if it is empty
	if empty, err := afero.IsEmpty(m.appFS, dirPathOld); err != nil {
		return err
	} else if empty {
		return m.appFS.RemoveAll(dirPathOld)
	}
	return nil
}

func (m *manager) cleanEmptyParentDirs(name string) error {
	// clean up any empty parent directory paths
	log.Debug("Removing empty parent directories, if any")
	parentDir := name
	for parentDir != "." {
		parentDir = filepath.Dir(parentDir)
		parentPath := str.AppendToPath(m.environmentsPath, parentDir)

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

func (m *manager) generateOverrideData() []byte {
	const (
		relBaseLibsonnetPath = "../" + baseLibsonnetFile
	)

	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("local base = import \"%s\";\n", relBaseLibsonnetPath))
	buf.WriteString(fmt.Sprintf("local k = import \"%s\";\n\n", lib.ExtensionsLibFilename))
	buf.WriteString("base + {\n")
	buf.WriteString("  // Insert user-specified overrides here. For example if a component is named \"nginx-deployment\", you might have something like:\n")
	buf.WriteString("  //   \"nginx-deployment\"+: k.deployment.mixin.metadata.labels({foo: \"bar\"})\n")
	buf.WriteString("}\n")
	return buf.Bytes()
}

func (m *manager) generateParamsData() []byte {
	const (
		relComponentParamsPath = "../../" + componentsDir + "/" + paramsFileName
	)

	return []byte(`local params = import "` + relComponentParamsPath + `";
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

func (m *manager) environmentExists(name string) (bool, error) {
	appSpec, err := m.AppSpec()
	if err != nil {
		return false, err
	}

	_, ok := appSpec.GetEnvironmentSpec(name)
	return ok, nil
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
