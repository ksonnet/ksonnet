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
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"sync"

	"github.com/ghodss/yaml"
	"github.com/ksonnet/ksonnet/pkg/lib"
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
	// libPath caches ksonnet lib paths after generation / validation
	libPaths map[string]string

	config    *Spec
	overrides *Override

	mu sync.Mutex

	load   func() error
	loaded bool
}

var _ App = (*baseApp)(nil)

// Opt is a constructor option for baseApp
type Opt func(*baseApp)

// OptLibUpdater returns an option for setting a KSLibUpdater on an App010
func OptLibUpdater(libUpdater KSLibUpdater) Opt {
	return func(a *baseApp) {
		a.libUpdater = libUpdater
	}
}

// optLoadFn overrides baseApp's load function, useful when testing.
func optLoadFn(loadFn func() error) Opt {
	return func(a *baseApp) {
		a.load = loadFn
	}
}

// optNopLoader overrides baseApp's loader to do nothing. (NOOP)
func optNoopLoader() Opt {
	return func(a *baseApp) {
		a.load = func() error { return nil }
	}
}

// NewBaseApp constructs a baseApp, a container of dependencies and configuration describing
// a ksonnet project.
func NewBaseApp(fs afero.Fs, root string, httpClient *http.Client, opts ...Opt) *baseApp {
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
		libPaths: make(map[string]string),
	}
	ba.load = ba.doLoad

	for _, optFn := range opts {
		optFn(ba)
	}

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
	if err := write(ba.fs, ba.root, ba.config); err != nil {
		return errors.Wrap(err, "serializing configuration")
	}

	if err := removeOverride(ba.fs, ba.root); err != nil {
		return errors.Wrap(err, "clean overrides")
	}

	if ba.overrides.IsDefined() {
		return saveOverride(defaultYAMLEncoder, ba.fs, ba.root, ba.overrides)
	}

	return nil
}

func (ba *baseApp) doLoad() error {
	ba.mu.Lock()
	defer ba.mu.Unlock()

	config, err := read(ba.fs, ba.root)
	if err != nil {
		return err
	}

	overrides, err := readOverrides(ba.fs, ba.root)
	if err != nil {
		return errors.Wrap(err, "reading app overrides")
	}

	if overrides == nil {
		overrides = newOverride()
	}

	ba.config = config
	ba.overrides = overrides
	ba.loaded = true

	return nil
}

func (ba *baseApp) AddRegistry(newReg *RegistryConfig, isOverride bool) error {
	if err := ba.load(); err != nil {
		return errors.Wrap(err, "load configuration")
	}

	if newReg.Name == "" {
		return ErrRegistryNameInvalid
	}

	var regMap = ba.config.Registries
	if isOverride {
		if ba.overrides == nil {
			ba.overrides = newOverride()
		}
		regMap = ba.overrides.Registries
	}

	_, exists := regMap[newReg.Name]
	if exists {
		return ErrRegistryExists
	}

	regMap[newReg.Name] = newReg

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
	if !ba.loaded {
		if err := ba.load(); err != nil {
			return nil, errors.Wrap(err, "load configuration")
		}
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
		return combined
	case hasOverride:
		e := deepCopyEnvironmentConfig(*override)
		return e
	default:
		return nil
	}
}

// Environments returns all environment specs, merged with any corresponding overrides.
// Note overrides cannot override environment libraries.
func (ba *baseApp) Environments() (EnvironmentConfigs, error) {
	if !ba.loaded {
		if err := ba.load(); err != nil {
			return nil, err
		}
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

// AddEnvironment adds an environment spec to the app spec. If the spec already exists,
// it is overwritten.
func (ba *baseApp) AddEnvironment(newEnv *EnvironmentConfig, k8sSpecFlag string, isOverride bool) error {
	log.WithFields(log.Fields{
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

	if err := ba.load(); err != nil {
		return errors.Wrap(err, "load configuration")
	}

	if k8sSpecFlag != "" {
		ver, err := ba.libUpdater.UpdateKSLib(k8sSpecFlag, app010LibPath(ba.root))
		if err != nil {
			return err
		}

		newEnv.KubernetesVersion = ver
	}

	var envMap = ba.config.Environments
	if isOverride {
		if ba.overrides == nil {
			ba.overrides = newOverride()
		}
		envMap = ba.overrides.Environments
	}

	envMap[newEnv.Name] = newEnv
	return ba.save()
}

// Libraries returns application libraries.
func (ba *baseApp) Libraries() (LibraryConfigs, error) {
	if !ba.loaded {
		if err := ba.load(); err != nil {
			return nil, errors.Wrap(err, "load configuration")
		}
	}

	return ba.config.Libraries, nil
}

// Registries returns application registries.
func (ba *baseApp) Registries() (RegistryConfigs, error) {
	if !ba.loaded {
		if err := ba.load(); err != nil {
			return nil, errors.Wrap(err, "load configuration")
		}
	}

	registries := RegistryConfigs{}

	for k, v := range ba.config.Registries {
		registries[k] = v
	}

	for k, v := range ba.overrides.Registries {
		registries[k] = v
	}

	return registries, nil
}

// RemoveEnvironment removes an environment.
func (ba *baseApp) RemoveEnvironment(envName string, override bool) error {
	if err := ba.load(); err != nil {
		return errors.Wrap(err, "load configuration")
	}

	envMap := ba.config.Environments
	if override {
		envMap = ba.overrides.Environments
	}

	if _, ok := envMap[envName]; !ok {
		return errors.Errorf("environment %q does not exist", envName)
	}

	delete(envMap, envName)

	return ba.save()
}

// RenameEnvironment renames environments.
func (ba *baseApp) RenameEnvironment(from, to string, override bool) error {
	if err := ba.load(); err != nil {
		return errors.Wrap(err, "load configuration")
	}

	envMap := ba.config.Environments
	if override {
		envMap = ba.overrides.Environments
	}

	if _, ok := envMap[from]; !ok {
		return errors.Errorf("environment %q does not exist", from)
	}
	envMap[to] = envMap[from]
	envMap[to].Path = to
	delete(envMap, from)

	if err := moveEnvironment(ba.fs, ba.root, from, to); err != nil {
		return err
	}

	return ba.save()
}

// UpdateTargets updates the list of targets. Note this overrwrite any existing targets.
func (ba *baseApp) UpdateTargets(envName string, targets []string) error {
	spec, err := ba.Environment(envName)
	if err != nil {
		return err
	}

	spec.Targets = targets

	isOverride := ba.IsEnvOverride(envName)
	return errors.Wrap(ba.AddEnvironment(spec, "", isOverride), "update targets")
}

// LibPath returns the lib path for an env environment.
func (ba *baseApp) LibPath(envName string) (string, error) {
	if lp, ok := ba.libPaths[envName]; ok {
		return lp, nil
	}

	env, err := ba.Environment(envName)
	if err != nil {
		return "", err
	}

	ver := fmt.Sprintf("version:%s", env.KubernetesVersion)
	lm, err := lib.NewManager(ver, ba.fs, app010LibPath(ba.root), ba.httpClient)
	if err != nil {
		return "", err
	}

	lp, err := lm.GetLibPath()
	if err != nil {
		return "", err
	}

	ba.checkKsonnetLib(lp)

	ba.libPaths[envName] = lp
	return lp, nil
}

// TODO move this to migrations somewhere
func (ba *baseApp) checkKsonnetLib(lp string) {
	libRoot := filepath.Join(ba.Root(), LibDirName, "ksonnet-lib")
	if !strings.HasPrefix(lp, libRoot) {
		log.Warnf("ksonnet has moved ksonnet-lib paths to %q. The current location of "+
			"of your existing ksonnet-libs can be automatically moved by ksonnet with `ks upgrade`",
			libRoot)
	}
}

// Upgrade upgrades an application (app.yaml) to the current version.
func (ba *baseApp) Upgrade(bool) error {
	if ba == nil {
		return errors.New("nil receiver")
	}
	if ba.config == nil {
		return errors.New("nil configuration")
	}
	if ba.fs == nil {
		return errors.New("nil fs interface")
	}

	appConfig, err := afero.ReadFile(ba.fs, specPath(ba.root))
	if err != nil {
		return err
	}

	var base specBase

	err = yaml.Unmarshal(appConfig, &base)
	if err != nil {
		return err
	}

	if base.APIVersion.String() == DefaultAPIVersion {
		// Nothing to do, schema on disk is already latest.
		return nil
	}

	if err := write(ba.fs, ba.root, ba.config); err != nil {
		return err
	}

	// TODO handle override upgrades

	return nil
}

// IsEnvOverride returns whether the specified environment has overriding configuration
func (ba *baseApp) IsEnvOverride(name string) bool {
	if ba.overrides == nil {
		return false
	}
	_, ok := ba.overrides.Environments[name]
	return ok
}

// IsRegistryOverride returns whether the specified registry has overriding configuration
func (ba *baseApp) IsRegistryOverride(name string) bool {
	if ba.overrides == nil {
		return false
	}
	_, ok := ba.overrides.Registries[name]
	return ok
}
