// Copyright 2018 The kubecfg authors
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

package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ksonnet/ksonnet/metadata/lib"

	"github.com/pkg/errors"
	"github.com/spf13/afero"
)

// App010 is a ksonnet 0.1.0 application.
type App010 struct {
	*baseApp
}

var _ App = (*App010)(nil)

// NewApp010 creates an App010 instance.
func NewApp010(fs afero.Fs, root string) *App010 {
	ba := newBaseApp(fs, root)

	a := &App010{
		baseApp: ba,
	}

	return a
}

// AddEnvironment adds an environment spec to the app spec. If the spec already exists,
// it is overwritten.
func (a *App010) AddEnvironment(name, k8sSpecFlag string, newEnv *EnvironmentSpec) error {
	spec, err := a.load()
	if err != nil {
		return err
	}

	spec.Environments[name] = newEnv

	if k8sSpecFlag != "" {
		ver, err := LibUpdater(a.fs, k8sSpecFlag, app010LibPath(a.root), true)
		if err != nil {
			return err
		}

		spec.Environments[name].KubernetesVersion = ver
	}

	return a.save(spec)
}

// Environment returns the spec for an environment.
func (a *App010) Environment(name string) (*EnvironmentSpec, error) {
	spec, err := a.load()
	if err != nil {
		return nil, err
	}

	s, ok := spec.Environments[name]
	if !ok {
		return nil, errors.Errorf("environment %q was not found", name)
	}

	return s, nil
}

// Environments returns all environment specs.
func (a *App010) Environments() (EnvironmentSpecs, error) {
	spec, err := a.load()
	if err != nil {
		return nil, err
	}

	return spec.Environments, nil
}

// Init initializes the App.
func (a *App010) Init() error {
	// check to see if there are spec.json files.

	legacyEnvs, err := a.findLegacySpec()
	if err != nil {
		return err
	}

	if len(legacyEnvs) == 0 {
		return nil
	}

	msg := "Your application's apiVersion is 0.1.0, but legacy environment declarations " +
		"where found in environments: %s. In order to proceed, you will have to run `ks upgrade` to " +
		"upgrade your application. <see url>"

	return errors.Errorf(msg, strings.Join(legacyEnvs, ", "))
}

// LibPath returns the lib path for an env environment.
func (a *App010) LibPath(envName string) (string, error) {
	env, err := a.Environment(envName)
	if err != nil {
		return "", err
	}

	ver := fmt.Sprintf("version:%s", env.KubernetesVersion)
	lm, err := lib.NewManager(ver, a.fs, app010LibPath(a.root))
	if err != nil {
		return "", err
	}

	return lm.GetLibPath(true)
}

// Libraries returns application libraries.
func (a *App010) Libraries() (LibraryRefSpecs, error) {
	spec, err := a.load()
	if err != nil {
		return nil, err
	}

	return spec.Libraries, nil
}

// Registries returns application registries.
func (a *App010) Registries() (RegistryRefSpecs, error) {
	spec, err := a.load()
	if err != nil {
		return nil, err
	}

	return spec.Registries, nil
}

// RemoveEnvironment removes an environment.
func (a *App010) RemoveEnvironment(envName string) error {
	spec, err := a.load()
	if err != nil {
		return err
	}
	delete(spec.Environments, envName)
	return a.save(spec)
}

// RenameEnvironment renames environments.
func (a *App010) RenameEnvironment(from, to string) error {
	if err := moveEnvironment(a.fs, a.root, from, to); err != nil {
		return err
	}

	spec, err := a.load()
	if err != nil {
		return err
	}

	spec.Environments[to] = spec.Environments[from]
	delete(spec.Environments, from)

	spec.Environments[to].Path = to

	return a.save(spec)
}

// UpdateTargets updates the list of targets for a 0.1.0 application.
func (a *App010) UpdateTargets(envName string, targets []string) error {
	spec, err := a.Environment(envName)
	if err != nil {
		return err
	}

	spec.Targets = targets

	return errors.Wrap(a.AddEnvironment(envName, "", spec), "update targets")
}

// Upgrade upgrades the app to the latest apiVersion.
func (a *App010) Upgrade(dryRun bool) error {
	return nil
}

func (a *App010) findLegacySpec() ([]string, error) {
	var found []string

	envPath := filepath.Join(a.root, EnvironmentDirName)
	err := afero.Walk(a.fs, envPath, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if fi.IsDir() {
			return nil
		}

		if fi.Name() == app001specJSON {
			envName := strings.TrimPrefix(path, envPath+"/")
			envName = strings.TrimSuffix(envName, "/"+app001specJSON)
			found = append(found, envName)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return found, nil
}
