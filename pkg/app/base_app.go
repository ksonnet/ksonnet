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

package app

import (
	"net/http"
	"path/filepath"
	"sync"

	"github.com/ghodss/yaml"
	stringutils "github.com/ksonnet/ksonnet/pkg/util/strings"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

type baseApp struct {
	root       string
	fs         afero.Fs
	httpClient *http.Client

	// LibUpdater updates ksonnet lib versions.
	libUpdater KSLibUpdater

	config    *Spec
	overrides *Override

	mu sync.Mutex

	load func() error
}

func newBaseApp(fs afero.Fs, root string, httpClient *http.Client) *baseApp {
	ba := &baseApp{
		fs:         fs,
		httpClient: httpClient,
		libUpdater: ksLibUpdater{
			fs:         fs,
			httpClient: httpClient,
		},
		root:   root,
		config: &Spec{},
		overrides: &Override{
			Environments: EnvironmentConfigs{},
			Registries:   RegistryConfigs{},
		},
	}
	ba.load = ba.doLoad
	return ba
}

func (ba *baseApp) CurrentEnvironment() string {
	currentPath := filepath.Join(ba.root, currentEnvName)
	data, err := afero.ReadFile(ba.fs, currentPath)
	if err != nil {
		return ""
	}

	return string(data)
}

func (ba *baseApp) SetCurrentEnvironment(name string) error {
	envs, err := ba.Environments()
	if err != nil {
		return errors.Wrap(err, "loading environments")
	}

	var envNames []string
	for _, env := range envs {
		envNames = append(envNames, env.Name)
	}

	if !stringutils.InSlice(name, envNames) {
		return errors.Errorf("environment %q does not exist", name)
	}

	currentPath := filepath.Join(ba.root, currentEnvName)
	return afero.WriteFile(ba.fs, currentPath, []byte(name), DefaultFilePermissions)
}

func (ba *baseApp) configPath() string {
	return filepath.Join(ba.root, "app.yaml")
}

func (ba *baseApp) overridePath() string {
	return filepath.Join(ba.root, "app.override.yaml")
}

func (ba *baseApp) save() error {
	log := log.WithField("action", "baseApp.save")

	ba.mu.Lock()
	defer ba.mu.Unlock()

	if ba.config == nil {
		return errors.Errorf("cannot save nil app configuration")
	}

	// Signal we have converted to new app version
	ba.config.APIVersion = DefaultAPIVersion
	log.Debugf("saving app version %v", ba.config.APIVersion)

	configData, err := yaml.Marshal(ba.config)
	if err != nil {
		return errors.Wrap(err, "convert application configuration to YAML")
	}

	if err = afero.WriteFile(ba.fs, ba.configPath(), configData, DefaultFilePermissions); err != nil {
		return errors.Wrapf(err, "write %s", ba.configPath())
	}

	if err = removeOverride(ba.fs, ba.root); err != nil {
		return errors.Wrap(err, "clean overrides")
	}

	if ba.overrides.IsDefined() {
		return SaveOverride(defaultYAMLEncoder, ba.fs, ba.root, ba.overrides)
	}

	return nil
}

func (ba *baseApp) doLoad() error {
	ba.mu.Lock()
	defer ba.mu.Unlock()

	configData, err := afero.ReadFile(ba.fs, ba.configPath())
	if err != nil {
		return errors.Wrapf(err, "read %s", ba.configPath())
	}

	var config Spec
	if err = yaml.Unmarshal(configData, &config); err != nil {
		return errors.Wrapf(err, "unmarshal application YAML config")
	}

	exists, err := afero.Exists(ba.fs, ba.overridePath())
	if err != nil {
		return err
	}

	if len(config.Environments) == 0 {
		config.Environments = EnvironmentConfigs{}
	}

	if len(config.Registries) == 0 {
		config.Registries = RegistryConfigs{}
	}

	override := Override{
		Environments: EnvironmentConfigs{},
		Registries:   RegistryConfigs{},
	}
	if exists {
		overrideData, err := afero.ReadFile(ba.fs, ba.overridePath())
		if err != nil {
			return errors.Wrapf(err, "read %s", ba.overridePath())
		}
		if err = yaml.Unmarshal(overrideData, &override); err != nil {
			return errors.Wrapf(err, "unmarshal override YAML config")
		}

		if err = override.Validate(); err != nil {
			return errors.Wrap(err, "validating override")
		}

		if len(override.Environments) == 0 {
			override.Environments = EnvironmentConfigs{}
		}

		if len(override.Registries) == 0 {
			override.Registries = RegistryConfigs{}
		}

		for k := range override.Registries {
			override.Registries[k].isOverride = true
		}

		for k := range override.Environments {
			override.Environments[k].isOverride = true
		}

	}

	ba.overrides = &override
	ba.config = &config

	return ba.config.validate()
}

func (ba *baseApp) AddRegistry(newReg *RegistryConfig, isOverride bool) error {
	if err := ba.load(); err != nil {
		return errors.Wrap(err, "load configuration")
	}

	if newReg.Name == "" {
		return ErrRegistryNameInvalid
	}

	if isOverride {
		_, exists := ba.overrides.Registries[newReg.Name]
		if exists {
			return ErrRegistryExists
		}

		newReg.isOverride = true

		ba.overrides.Registries[newReg.Name] = newReg
		return ba.save()
	}

	_, exists := ba.config.Registries[newReg.Name]
	if exists {
		return ErrRegistryExists
	}

	newReg.isOverride = false
	ba.config.Registries[newReg.Name] = newReg

	return ba.save()
}

// libKeyByDesc scans for a library in the referenced LibraryConfigs map.
// Matching is by registry/name or name if registry is not provided.
// Versions are ignored.
// Complexity: O(n)
// Returns key of matching library, or empty string if no match is found.
func libKeyByDesc(desc LibraryConfig, libs LibraryConfigs) string {
	for k, v := range libs {
		if desc.Name != v.Name {
			continue
		}
		if desc.Registry == "" {
			return k
		}
		if desc.Registry == v.Registry {
			return k
		}
	}
	return ""
}

// UpdateLib adds or updates a library reference.
// env is optional - if provided the reference is scoped under the environment,
// otherwise it is globally scoped.
// If spec if nil, the library reference will be removed.
// Returns the previous reference for the named library, if one existed.
func (ba *baseApp) UpdateLib(id string, env string, libSpec *LibraryConfig) (*LibraryConfig, error) {
	if err := ba.load(); err != nil {
		return nil, errors.Wrap(err, "load configuration")
	}

	if ba.config == nil {
		return nil, errors.Errorf("invalid app - configuration is nil")
	}

	desc, err := parseLibraryConfig(id)
	if err != nil {
		return nil, errors.Wrapf(err, "parsing id: %s", id)
	}

	if libSpec != nil && libSpec.Name != desc.Name {
		return nil, errors.Errorf("library name mismatch: %v vs %v", libSpec.Name, desc.Name)
	}

	// TODO support app overrides

	var oldSpec *LibraryConfig
	switch env {
	case "":
		// Globally scoped
		oldSpec, err = updateLibInScope(desc, libSpec, &ba.config.Libraries)
		if err != nil {
			return nil, errors.Wrap(err, "updating library in global scope")
		}
	default:
		// Scoped by environment
		e, ok := ba.config.GetEnvironmentConfig(env)
		if !ok {
			return nil, errors.Errorf("invalid environment: %v", env)
		}

		oldSpec, err = updateLibInScope(desc, libSpec, &e.Libraries)
		if err != nil {
			return nil, errors.Wrap(err, "updating library in environment scope")
		}

		if err := ba.config.UpdateEnvironmentConfig(env, e); err != nil {
			return nil, errors.Wrapf(err, "updating environment %v", env)
		}
	}

	return oldSpec, ba.save()
}

func updateLibInScope(desc LibraryConfig, libSpec *LibraryConfig, scope *LibraryConfigs) (*LibraryConfig, error) {
	if scope == nil {
		return nil, errors.New("nil scope")
	}
	if *scope == nil {
		*scope = LibraryConfigs{}
	}

	oldSpecKey := libKeyByDesc(desc, *scope)
	oldSpec := (*scope)[oldSpecKey]
	if libSpec == nil {
		if oldSpecKey == "" || oldSpec == nil {
			return nil, errors.Errorf("package not found: %v", desc)
		}
		delete(*scope, oldSpecKey)
	} else {
		newKey := qualifyLibName(libSpec.Registry, libSpec.Name)
		(*scope)[newKey] = libSpec
	}

	return oldSpec, nil
}

// UpdateRegistry updates a registry spec and persists in app[.override].yaml
func (ba *baseApp) UpdateRegistry(spec *RegistryConfig) error {
	if err := ba.load(); err != nil {
		return errors.Wrap(err, "load configuration")
	}

	if spec.Name == "" {
		return ErrRegistryNameInvalid
	}

	// Figure out where the registry is defined (app or override)
	var ok, okOverride bool
	if ba.config != nil {
		_, ok = ba.config.Registries[spec.Name]
	}
	if ba.overrides != nil {
		_, okOverride = ba.overrides.Registries[spec.Name]
	}

	if !ok && !okOverride {
		return errors.Errorf("registry not found: %v", spec.Name)
	}

	if ok && okOverride {
		return errors.Errorf("registry %v found in both app.yaml and app.override.yaml", spec.Name)
	}

	if ok {
		ba.config.Registries[spec.Name] = spec
	} else {
		ba.overrides.Registries[spec.Name] = spec
	}

	return ba.save()
}

func (ba *baseApp) Fs() afero.Fs {
	return ba.fs
}

func (ba *baseApp) HTTPClient() *http.Client {
	return ba.httpClient
}

func (ba *baseApp) Root() string {
	return ba.root
}

func (ba *baseApp) EnvironmentParams(envName string) (string, error) {
	if envName == "" {
		return "", errors.New("environment name is blank")
	}
	envParamsPath := filepath.Join(ba.Root(), EnvironmentDirName, envName, "params.libsonnet")
	b, err := afero.ReadFile(ba.Fs(), envParamsPath)
	if err != nil {
		return "", errors.Wrapf(err, "unable to read params for environment %s", envName)
	}

	return string(b), nil
}

func (ba *baseApp) VendorPath() string {
	return filepath.Join(ba.Root(), "vendor")
}

// Environment returns the spec for an environment.
func (ba *baseApp) Environment(name string) (*EnvironmentConfig, error) {
	if err := ba.load(); err != nil {
		return nil, errors.Wrap(err, "load configuration")
	}

	e := ba.mergedEnvironment(name)
	if e == nil {
		return nil, errors.Errorf("environment %q was not found", name)
	}
	return e, nil
}

func deepCopyLibraries(src LibraryConfigs) LibraryConfigs {
	if src == nil {
		return LibraryConfigs(nil)
	}

	lc := LibraryConfigs{}
	for k, v := range src {
		c := *v
		lc[k] = &c
	}
	return lc
}

func deepCopyEnvironmentConfig(src EnvironmentConfig) *EnvironmentConfig {
	e := src

	if src.Destination != nil {
		d := *src.Destination
		e.Destination = &d
	}
	if src.Targets != nil {
		t := make([]string, len(src.Targets))
		copy(t, src.Targets)
		e.Targets = t
	}
	if src.Libraries != nil {
		e.Libraries = deepCopyLibraries(src.Libraries)
	}

	return &e
}

// mergedEnvrionment returns a fresh copy of the named environment, merged with
// optional overrides if present. Note overrides cannot override environment-scoped library
// references.
// Returns nil if the envrionment is not present and non-nil in either primary configuration
// or overrides.
func (ba *baseApp) mergedEnvironment(name string) *EnvironmentConfig {
	var hasPrimary, hasOverride bool
	var primary, override *EnvironmentConfig

	if ba.config != nil {
		primary, hasPrimary = ba.config.Environments[name]
		if primary == nil {
			hasPrimary = false
		}
	}
	if ba.overrides != nil {
		override, hasOverride = ba.overrides.Environments[name]
		if override == nil {
			hasOverride = false
		}
	}

	switch {
	case hasPrimary && !hasOverride:
		e := deepCopyEnvironmentConfig(*primary)
		e.isOverride = false
		return e
	case hasPrimary && hasOverride:
		combined := deepCopyEnvironmentConfig(*primary)
		combined.Name = override.Name
		combined.KubernetesVersion = override.KubernetesVersion
		combined.Path = override.Path
		if override.Destination != nil {
			d := *override.Destination
			combined.Destination = &d
		}
		if override.Targets != nil {
			t := make([]string, len(override.Targets))
			copy(t, override.Targets)
			combined.Targets = t
		}
		combined.isOverride = true
		return combined
	case hasOverride:
		e := deepCopyEnvironmentConfig(*override)
		e.isOverride = true
		return e
	default:
		return nil
	}
}

// Environments returns all environment specs, merged with any corresponding overrides.
// Note overrides cannot override environment libraries.
func (ba *baseApp) Environments() (EnvironmentConfigs, error) {
	if err := ba.load(); err != nil {
		return nil, errors.Wrap(err, "load configuration")
	}

	// Build merged list of keys
	environments := EnvironmentConfigs{}
	if ba.config != nil {
		for k := range ba.config.Environments {
			environments[k] = nil
		}
	}
	if ba.overrides != nil {
		for k := range ba.overrides.Environments {
			environments[k] = nil
		}
	}

	for k := range environments {
		e := ba.mergedEnvironment(k)
		if e == nil {
			delete(environments, k)
			continue
		}

		environments[k] = e
	}

	return environments, nil
}
