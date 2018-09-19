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
	"net/http"
	"os"
	"path/filepath"
	"sort"

	"github.com/pkg/errors"
	"github.com/spf13/afero"
)

const (

	// appYamlName is the name for the app configuration.
	appYamlName = "app.yaml"

	// overrideYamlName is the name for the app overrides.
	overrideYamlName = "app.override.yaml"

	// EnvironmentDirName is the directory name for environments.
	EnvironmentDirName = "environments"

	// LibDirName is the directory name for libraries.
	LibDirName = "lib"

	// currentEnvName is the file which selects the current environment.
	currentEnvName = ".ks_environment"
)

var (
	// DefaultFilePermissions are the default permissions for a file.
	DefaultFilePermissions = os.FileMode(0644)
	// DefaultFolderPermissions are the default permissions for a folder.
	DefaultFolderPermissions = os.FileMode(0755)
)

// App is a ksonnet application.
type App interface {
	// AddEnvironment adds an environment.
	AddEnvironment(spec *EnvironmentConfig, k8sSpecFlag string, isOverride bool) error
	// AddRegistry adds a registry.
	AddRegistry(spec *RegistryConfig, isOverride bool) error
	// CurrentEnvironment returns the current environment name or an empty string.
	CurrentEnvironment() string
	// Environment finds an environment by name.
	Environment(name string) (*EnvironmentConfig, error)
	// Environments returns all environments.
	Environments() (EnvironmentConfigs, error)
	// EnvironmentParams returns params for an environment.
	EnvironmentParams(name string) (string, error)
	// Fs is the app's afero Fs.
	Fs() afero.Fs
	// HTTPClient is the app's http client
	HTTPClient() *http.Client
	// IsEnvOverride returns whether the specified environment has overriding configuration
	IsEnvOverride(name string) bool
	// IsRegistryOverride returns whether the specified registry has overriding configuration
	IsRegistryOverride(name string) bool
	// LibPath returns the path of the lib for an environment.
	LibPath(envName string) (string, error)
	// Libraries returns all environments.
	Libraries() (LibraryConfigs, error)
	// Registries returns all registries.
	Registries() (RegistryConfigs, error)
	// RemoveEnvironment removes an environment from the main configuration or an override.
	RemoveEnvironment(name string, override bool) error
	// RenameEnvironment renames an environment in the main configuration or an override.
	RenameEnvironment(from, to string, override bool) error
	// Root returns the root path of the application.
	Root() string
	// SetCurrentEnvironment sets the current environment.
	SetCurrentEnvironment(name string) error
	// UpdateTargets sets the targets for an environment.
	UpdateTargets(envName string, targets []string, isOverride bool) error
	// UpdateLib adds, updates or removes a library reference.
	// env is optional - if provided the reference is scoped under the environment,
	// otherwise it is globally scoped.
	// If spec if nil, the library reference will be removed.
	// Returns the previous reference for the named library, if one existed.
	UpdateLib(name string, env string, spec *LibraryConfig) (*LibraryConfig, error)
	// UpdateRegistry updates a registry.
	UpdateRegistry(spec *RegistryConfig) error
	// Upgrade upgrades an application (app.yaml) to the current version.
	Upgrade(bool) error
	// VendorPath returns the root of the vendor path.
	VendorPath() string
}

// Load loads the application configuration.
func Load(fs afero.Fs, httpClient *http.Client, appRoot string) (App, error) {
	if fs == nil {
		return nil, errors.New("nil fs interface")
	}

	_, err := fs.Stat(specPath(appRoot))
	if os.IsNotExist(err) {
		// During `ks init`, app.yaml will not yet exist - generate a new one.
		return NewBaseApp(fs, appRoot, httpClient), nil
	}
	if err != nil {
		return nil, errors.Wrap(err, "checking existence of app configuration")
	}

	a := NewBaseApp(fs, appRoot, httpClient)
	if err := a.doLoad(); err != nil {
		return nil, errors.Wrap(err, "reading app configuration")
	}

	return a, nil
}

func app010LibPath(root string) string {
	return filepath.Join(root, LibDirName)
}

// StubUpdateLibData always returns no error.
func StubUpdateLibData(fs afero.Fs, k8sSpecFlag, libPath string) (string, error) {
	return "v1.8.7", nil
}

func moveEnvironment(fs afero.Fs, root, from, to string) error {
	toPath := filepath.Join(root, EnvironmentDirName, to)

	exists, err := afero.Exists(fs, filepath.Join(toPath, "main.jsonnet"))
	if err != nil {
		return err
	}

	if exists {
		return errors.Errorf("unable to rename %q because %q exists", from, to)
	}

	fromPath := filepath.Join(root, EnvironmentDirName, from)
	exists, err = afero.Exists(fs, fromPath)
	if err != nil {
		return err
	}

	if !exists {
		return errors.Errorf("environment %q does not exist", from)
	}

	// create to directory
	if err = fs.MkdirAll(toPath, DefaultFolderPermissions); err != nil {
		return err
	}

	fis, err := afero.ReadDir(fs, fromPath)
	if err != nil {
		return err
	}

	for _, fi := range fis {
		if fi.IsDir() && fi.Name() != ".metadata" {
			continue
		}

		oldPath := filepath.Join(fromPath, fi.Name())
		newPath := filepath.Join(toPath, fi.Name())
		if err := fs.Rename(oldPath, newPath); err != nil {
			return err
		}
	}

	return cleanEnv(fs, root)
}

func cleanEnv(fs afero.Fs, root string) error {
	var dirs []string

	envDir := filepath.Join(root, EnvironmentDirName)
	err := afero.Walk(fs, envDir, func(path string, fi os.FileInfo, err error) error {
		if !fi.IsDir() {
			return nil
		}

		dirs = append(dirs, path)
		return nil
	})

	if err != nil {
		return err
	}

	sort.Sort(sort.Reverse(sort.StringSlice(dirs)))

	for _, dir := range dirs {
		fis, err := afero.ReadDir(fs, dir)
		if err != nil {
			return err
		}

		if len(fis) == 0 {
			if err := fs.RemoveAll(dir); err != nil {
				return err
			}
		}
	}

	return nil
}

// FindRoot finds a ksonnet app.yaml in the current directory or its ancestors.
func FindRoot(fs afero.Fs, cwd string) (string, error) {
	prev := cwd

	for {
		path := filepath.Join(cwd, appYamlName)
		exists, err := afero.Exists(fs, path)
		if err != nil {
			return "", err
		}

		if exists {
			return cwd, nil
		}

		cwd, err = filepath.Abs(filepath.Join(cwd, ".."))
		if err != nil {
			return "", err
		}

		if cwd == prev {
			return "", errors.Errorf("unable to find ksonnet project")
		}

		prev = cwd
	}
}
