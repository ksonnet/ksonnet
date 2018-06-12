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

package actions

import (
	"io"
	"os"

	"github.com/ksonnet/ksonnet/pkg/app"
	"github.com/ksonnet/ksonnet/pkg/registry"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// RunRegistryUpdate runs `env list`
func RunRegistryUpdate(m map[string]interface{}) error {
	ru, err := NewRegistryUpdate(m)
	if err != nil {
		return err
	}

	ol := newOptionLoader(m)
	name := ol.LoadString(OptionName)
	version := ol.LoadString(OptionVersion)

	if ol.err != nil {
		return ol.err
	}

	return ru.run(name, version)
}

type LocateFn func(app.App, *app.RegistryRefSpec) (registry.Updater, error)

// RegistryUpdate lists available registries
type RegistryUpdate struct {
	app      app.App
	listFn   func(ksApp app.App) ([]registry.Registry, error)
	locateFn LocateFn
	out      io.Writer
}

// NewRegistryUpdate creates an instance of RegistryUpdate
func NewRegistryUpdate(m map[string]interface{}) (*RegistryUpdate, error) {
	ol := newOptionLoader(m)

	ru := &RegistryUpdate{
		app: ol.LoadApp(),

		listFn:   registry.List,
		locateFn: defaultLocate,
		out:      os.Stdout,
	}

	if ol.err != nil {
		return nil, ol.err
	}

	return ru, nil
}

// defaultLocate passes-through to registry.Locate, but constrains the interface
// to just `registry.Updater`. The concrete type of registry.Updater is determined
// by the `spec` argument.
func defaultLocate(ksApp app.App, spec *app.RegistryRefSpec) (registry.Updater, error) {
	return registry.Locate(ksApp, spec)
}

// resolveUpdateSet returns a list of registries (by name) to update, based on user input.
// If a name was given, that registry will be the sole member of the updateSet.
// Otherwise, a list of all currently configured registries will be returned.
func (ru *RegistryUpdate) resolveUpdateSet(name string) ([]string, error) {
	if ru == nil {
		return nil, errors.Errorf("nil receiver")
	}

	// Empty registry name == all
	updateSet := make([]string, 0)

	specs, err := ru.app.Registries()
	if err != nil {
		return nil, errors.Wrap(err, "error retrieving configured registries")
	}

	switch name {
	case "":
		// The user asked to update all registries
		for regName := range specs {
			updateSet = append(updateSet, regName)
		}

	default:
		// The user asked to update a specific registry -
		// ensure it is valid.
		if _, ok := specs[name]; !ok {
			return nil, errors.Errorf("`unknown` registry: %v", name)
		}

		updateSet = append(updateSet, name)
	}
	return updateSet, nil
}

// verifyRegistryExists verifies that a registry of a given name is
// configured in the current application.
func (ru *RegistryUpdate) verifyRegistryExists(name string) (bool, error) {
	if ru == nil {
		return false, errors.Errorf("nil receiver")
	}

	if name == "" {
		return false, errors.Errorf("registry name required")
	}

	if ru.app == nil {
		return false, errors.Errorf("missing application")
	}

	// NOTE: app.Registries() does not currently cache app configuration
	specs, err := ru.app.Registries()
	if err != nil {
		return false, errors.Wrap(err, "error retrieving configured registries")
	}

	_, ok := specs[name]
	return ok, nil
}

// run runs the registry update command.
// Both name and version are optional.
// Empty name means all registries rather than a specific one.
// Empty version means try to use the latest version matching the registry's spec.
func (ru *RegistryUpdate) run(name string, version string) error {
	if ru == nil {
		return errors.Errorf("nil receiver")
	}

	// Figure our which registries to update.
	updateSet, err := ru.resolveUpdateSet(name)
	if err != nil {
		return errors.Wrap(err, "failed to resolve registry update set")
	}

	if len(updateSet) < 1 {
		return errors.Errorf("no registries to update")
	}

	log.Debugf("Updating registries: %v\n", updateSet)
	return doUpdate(ru.app, ru.locateFn, updateSet, version)
}

// doUpdate updates the provided registries. The optional version will be used if provided.
// Otherwise, latest versions matching the current registry specs will be used.
func doUpdate(app app.App, locateFn LocateFn, updateSet []string, version string) error {
	if app == nil {
		return errors.Errorf("missing application")
	}

	if locateFn == nil {
		return errors.Errorf("missing registry locator function")
	}

	if len(updateSet) == 0 {
		return errors.Errorf("nothing to update")
	}

	registries, err := app.Registries()
	if err != nil {
		return errors.Wrap(err, "failed to retrieve registries")
	}

	for _, name := range updateSet {
		// Resolve the registry by name
		rs, ok := registries[name]
		if !ok {
			return errors.Errorf("registry not found: %v", name)
		}

		log.Debugf("updating registry %v", name)
		regUpdate, err := locateFn(app, rs)
		if err != nil {
			return errors.Wrapf(err, "could not locate registry by spec: %v", rs.Name)
		}

		_, err = doUpdateRegistry(app, regUpdate, rs, version)
		if err != nil {
			return errors.Wrapf(err, "error updating registry: %v", rs.Name)
		}
	}

	return nil
}

// doUpdateRegisrty updates a single registry.
// `app` maps registries by name. It will be updated with the new version if a change has occured.
// `regUpdate` is an updatable registry. It will resolve the new version if `version` was not provided.
// `rs` is a registry spec representing the current name and version of the registry, prior to update.
// `version` is the optional desired version. When none is provided, the registry will attempt to resolve
//           to the latest version matching the registry specifier.
// Returns new version after update, and optional error.
func doUpdateRegistry(a app.App, regUpdate registry.Updater, rs *app.RegistryRefSpec, version string) (string, error) {
	if a == nil {
		return "", errors.Errorf("missing application")
	}

	if rs == nil {
		return "", errors.Errorf("nothing to update")
	}

	if version != "" {
		return "", errors.Errorf("TODO not implemented")
	}

	newVersion, err := regUpdate.Update(version)
	if err != nil {
		return "", errors.Wrapf(err, "update failed for registry: %v", rs.Name)
	}

	// If changed, update app.yaml to point to new version
	var oldVersion string
	if rs.GitVersion != nil {
		oldVersion = rs.GitVersion.CommitSHA
	}

	if oldVersion != newVersion {
		log.Debugf("[update] registry %v version updated from '%v' to '%v'",
			rs.Name, oldVersion, newVersion)

		// Make a new registryRefSpec. Create GitVersion even if there was none previously.
		newRS := *rs
		var newGitVersion app.GitVersionSpec
		if rs.GitVersion != nil {
			newGitVersion = *(rs.GitVersion)
		}
		newGitVersion.CommitSHA = newVersion
		newRS.GitVersion = &newGitVersion

		if err := a.UpdateRegistry(&newRS); err != nil {
			return "", errors.Wrapf(err, "error updating app registry pointer: %v", rs.Name)
		}
	} else {
		log.Debugf("[update] registry %v version unchanged: %v", rs.Name, newVersion)
		// TODO where does helm store its versions?
	}

	return newVersion, nil
}
