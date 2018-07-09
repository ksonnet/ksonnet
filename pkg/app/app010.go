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
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/ksonnet/ksonnet/pkg/lib"
	"github.com/sirupsen/logrus"

	"github.com/pkg/errors"
	"github.com/spf13/afero"
)

// App010 is a ksonnet 0.1.0 application.
type App010 struct {
	*baseApp

	out      io.Writer
	libPaths map[string]string
}

var _ App = (*App010)(nil)

// NewApp010 creates an App010 instance.
func NewApp010(fs afero.Fs, root string) *App010 {
	ba := newBaseApp(fs, root)

	a := &App010{
		baseApp: ba,
		out:     os.Stdout,

		libPaths: make(map[string]string),
	}

	return a
}

// AddEnvironment adds an environment spec to the app spec. If the spec already exists,
// it is overwritten.
func (a *App010) AddEnvironment(newEnv *EnvironmentConfig, k8sSpecFlag string, isOverride bool) error {
	logrus.WithFields(logrus.Fields{
		"k8s-spec-flag": k8sSpecFlag,
		"name":          newEnv.Name,
	}).Debug("adding environment")

	if newEnv == nil {
		return errors.Errorf("nil environment")
	}

	if newEnv.Name == "" {
		return errors.Errorf("invalid environment name")
	}

	if isOverride && len(newEnv.Libraries) > 0 {
		return errors.Errorf("library references not allowed in overrides")
	}

	if err := a.load(); err != nil {
		return errors.Wrap(err, "load configuration")
	}

	if k8sSpecFlag != "" {
		ver, err := LibUpdater(a.fs, k8sSpecFlag, app010LibPath(a.root))
		if err != nil {
			return err
		}

		newEnv.KubernetesVersion = ver
	}

	newEnv.isOverride = isOverride

	if isOverride {
		a.overrides.Environments[newEnv.Name] = newEnv
	} else {
		a.config.Environments[newEnv.Name] = newEnv
	}

	return a.save()
}

// CheckUpgrade initializes the App.
func (a *App010) CheckUpgrade() (bool, error) {
	if a == nil {
		return false, errors.Errorf("nil reciever")
	}

	var needUpgrade bool

	legacyLibs, err := a.checkUpgradeVendoredPackages()
	if err != nil {
		return false, err
	}

	if len(legacyLibs) > 0 {
		logrus.Warnf("Versioned packages stored in unversioned paths - please run `ks upgrade` to correct.")
		needUpgrade = true
	}

	// check to see if there are spec.json files.
	legacyEnvs, err := a.findLegacySpec()
	if err != nil {
		return false, err
	}

	if len(legacyEnvs) == 0 {
		return needUpgrade, nil
	}

	needUpgrade = true
	apiVersion := "0.1.0"
	if a.config != nil {
		apiVersion = a.config.APIVersion
	}
	logrus.Warnf("Your application's apiVersion is %s, but legacy environment declarations "+
		"were found in environments: %s. In order to proceed, you will have to run `ks upgrade` to "+
		"upgrade your application. <see url>", apiVersion, strings.Join(legacyEnvs, ", "))

	return needUpgrade, nil
}

// LibPath returns the lib path for an env environment.
func (a *App010) LibPath(envName string) (string, error) {
	if lp, ok := a.libPaths[envName]; ok {
		return lp, nil
	}

	env, err := a.Environment(envName)
	if err != nil {
		return "", err
	}

	ver := fmt.Sprintf("version:%s", env.KubernetesVersion)
	lm, err := lib.NewManager(ver, a.fs, app010LibPath(a.root))
	if err != nil {
		return "", err
	}

	lp, err := lm.GetLibPath()
	if err != nil {
		return "", err
	}

	a.checkKsonnetLib(lp)

	a.libPaths[envName] = lp
	return lp, nil
}

func (a *App010) checkKsonnetLib(lp string) {
	libRoot := filepath.Join(a.Root(), LibDirName, "ksonnet-lib")
	if !strings.HasPrefix(lp, libRoot) {
		logrus.Warnf("ksonnet has moved ksonnet-lib paths to %q. The current location of "+
			"of your existing ksonnet-libs can be automatically moved by ksonnet with `ks upgrade`",
			libRoot)
	}
}

// Libraries returns application libraries.
func (a *App010) Libraries() (LibraryConfigs, error) {
	if err := a.load(); err != nil {
		return nil, errors.Wrap(err, "load configuration")
	}

	return a.config.Libraries, nil
}

// Registries returns application registries.
func (a *App010) Registries() (RegistryConfigs, error) {
	if err := a.load(); err != nil {
		return nil, errors.Wrap(err, "load configuration")
	}

	registries := RegistryConfigs{}

	for k, v := range a.config.Registries {
		registries[k] = v
	}

	for k, v := range a.overrides.Registries {
		registries[k] = v
	}

	return registries, nil
}

// RemoveEnvironment removes an environment.
func (a *App010) RemoveEnvironment(envName string, override bool) error {
	if err := a.load(); err != nil {
		return errors.Wrap(err, "load configuration")
	}

	envMap := a.config.Environments
	if override {
		envMap = a.overrides.Environments
	}

	if _, ok := envMap[envName]; !ok {
		return errors.Errorf("environment %q does not exist", envName)
	}

	delete(envMap, envName)

	return a.save()
}

// RenameEnvironment renames environments.
func (a *App010) RenameEnvironment(from, to string, override bool) error {
	if err := a.load(); err != nil {
		return errors.Wrap(err, "load configuration")
	}

	envMap := a.config.Environments
	if override {
		envMap = a.overrides.Environments
	}

	if _, ok := envMap[from]; !ok {
		return errors.Errorf("environment %q does not exist", from)
	}
	envMap[to] = envMap[from]
	envMap[to].Path = to
	delete(envMap, from)

	if err := moveEnvironment(a.fs, a.root, from, to); err != nil {
		return err
	}

	return a.save()
}

// UpdateTargets updates the list of targets for a 0.1.0 application.
func (a *App010) UpdateTargets(envName string, targets []string) error {
	spec, err := a.Environment(envName)
	if err != nil {
		return err
	}

	spec.Targets = targets

	return errors.Wrap(a.AddEnvironment(spec, "", spec.isOverride), "update targets")
}

// Upgrade upgrades the app to the latest apiVersion.
func (a *App010) Upgrade(dryRun bool) error {
	if a == nil {
		return errors.Errorf("nil receiver")
	}
	if a.config == nil {
		return errors.Errorf("invalid app - config is nil")
	}
	a.config.APIVersion = "0.2.0"
	return a.save()
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

// Returns library configurations for the app, whether they are global or environment-scoped.
func (a *App010) allLibraries() (LibraryConfigs, error) {
	if a == nil {
		return nil, errors.Errorf("nil receiver")
	}

	combined := LibraryConfigs{}

	libs, err := a.Libraries()
	if err != nil {
		return nil, errors.Wrapf(err, "checking libraries")
	}
	for _, lib := range libs {
		combined[lib.Name] = lib
	}

	envs, err := a.Environments()
	if err != nil {
		return nil, errors.Wrapf(err, "checking environments")
	}
	for _, env := range envs {
		for _, lib := range env.Libraries {
			// NOTE We do not check for collisions at this time
			combined[lib.Name] = lib
		}
	}

	return combined, nil
}

// checkUpgradeVendoredPackages checks whether vendored packages need to be upgraded.
// Upgrades are necessary if a versioned package is stored in pre-ksonnet 0.12.0, unversioned directory.
func (a *App010) checkUpgradeVendoredPackages() ([]*LibraryConfig, error) {
	if a == nil {
		return nil, errors.Errorf("nil receiver")
	}
	fs := a.Fs()
	if fs == nil {
		return nil, errors.Errorf("nil filesystem interface")
	}

	combined, err := a.allLibraries()
	if err != nil {
		return nil, errors.Wrapf(err, "retrieving libraries")
	}

	results := make([]*LibraryConfig, 0)
	for _, l := range combined {
		if l.Version == "" {
			continue
		}

		path := filepath.Join(a.VendorPath(), l.Registry, l.Name)
		ok, err := afero.DirExists(fs, path)
		if err != nil {
			return nil, err
		}
		if ok {
			results = append(results, l)
		}
	}

	return results, nil
}
