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

package registry

import (
	"fmt"
	"strings"

	"github.com/ksonnet/ksonnet/pkg/app"
	"github.com/ksonnet/ksonnet/pkg/parts"
	"github.com/ksonnet/ksonnet/pkg/pkg"
	"github.com/ksonnet/ksonnet/pkg/prototype"
	"github.com/pkg/errors"
)

// InstalledChecker checks if a package is installed, based on app config.
type InstalledChecker interface {
	// IsInstalled checks whether a package is installed.
	// Supports fuzzy searches: Name, Name/Version, Registry/Name/Version, Registry/Version
	IsInstalled(d pkg.Descriptor) (bool, error)
}

// PackageManager is a package manager.
type PackageManager interface {
	Find(string) (pkg.Package, error)

	// Packages lists packages.
	Packages() ([]pkg.Package, error)

	// PackagesForEnv returns a list of Packages defined in the application, from the context
	// of the specified environment.
	PackagesForEnv(e *app.EnvironmentConfig) ([]pkg.Package, error)

	// RemotePackages returns a list of remote packages.
	RemotePackages() ([]pkg.Package, error)

	// Prototypes lists prototypes.
	Prototypes() (prototype.Prototypes, error)

	// PackageEnvironments returns a list of environments a package is installed in.
	PackageEnvironments(pkg pkg.Package) ([]*app.EnvironmentConfig, error)

	InstalledChecker
}

type registryConfigLister interface {
	Registries() (app.RegistryConfigs, error)
}

// packageManager is an implementation of PackageManager.
type packageManager struct {
	app app.App

	InstallChecker pkg.InstallChecker
	packagesFn     func() ([]pkg.Package, error)
	registriesFn   func() (map[string]SpecFetcher, error)
	resolverFn     func(name string) (LibrarySpecResolver, error)
	environmentsFn func() (app.EnvironmentConfigs, error)
}

var _ PackageManager = (*packageManager)(nil)

// NewPackageManager creates an instance of PackageManager.
func NewPackageManager(a app.App) PackageManager {
	pm := packageManager{
		app:            a,
		InstallChecker: &pkg.DefaultInstallChecker{App: a},
	}
	pm.packagesFn = pm.Packages
	pm.registriesFn = func() (map[string]SpecFetcher, error) {
		r, err := resolveRegistries(a)
		if err != nil {
			return nil, err
		}

		return registriesToSpecFetchers(r), nil
	}
	pm.resolverFn = func(name string) (LibrarySpecResolver, error) {
		r, err := resolveRegistry(a, name)
		if err != nil {
			return nil, err
		}
		return LibrarySpecResolver(r), nil
	}
	if a != nil {
		pm.environmentsFn = a.Environments
	} else {
		pm.environmentsFn = func() (app.EnvironmentConfigs, error) {
			return nil, errors.New("not implemented")
		}
	}
	return &pm
}

// resolveRegistries returns a list of registries from the provided app.
// (SpecFetcher is a subset of the Registry interface)
func resolveRegistries(a app.App) (map[string]Registry, error) {
	if a == nil {
		return nil, errors.New("nil app")
	}

	cfgs, err := a.Registries()
	if err != nil {
		return nil, err
	}

	result := make(map[string]Registry)
	for _, cfg := range cfgs {
		r, err := Locate(a, cfg)
		if err != nil {
			return nil, errors.Wrapf(err, "resolving registry: %v", cfg.Name)
		}
		result[cfg.Name] = r
	}

	return result, nil
}

// resolveRegistry returns the named registry from the provided app.
func resolveRegistry(a app.App, name string) (Registry, error) {
	if a == nil {
		return nil, errors.New("nil app")
	}

	all, err := resolveRegistries(a)
	if err != nil {
		return nil, err
	}

	r, ok := all[name]
	if !ok {
		return nil, errors.Errorf("registry not found: %s", name)
	}

	return r, nil
}

// Maps map[string]Registry -> map[string]SpecFectcher
func registriesToSpecFetchers(r map[string]Registry) map[string]SpecFetcher {
	result := make(map[string]SpecFetcher)
	for k, v := range r {
		result[k] = SpecFetcher(v)
	}
	return result
}

// Maps map[string]Registry -> map[string]Resolver
func registriesToResolvers(r map[string]Registry) map[string]LibrarySpecResolver {
	result := make(map[string]LibrarySpecResolver)
	for k, v := range r {
		result[k] = LibrarySpecResolver(v)
	}
	return result
}

// registryPrototcol returns the protocol for the named registry.
func registryProtocol(a registryConfigLister, name string) (proto Protocol, found bool) {
	regs, err := a.Registries()
	if err != nil {
		return ProtocolInvalid, false
	}

	r, ok := regs[name]
	if !ok {
		return ProtocolInvalid, false
	}
	return Protocol(r.Protocol), true
}

// Find finds a package by name. Package names have the format `<registry>/<library>@<version>`.
// Remote registries may be consulted if the package is not installed locally.
func (m *packageManager) Find(name string) (pkg.Package, error) {
	if m.app == nil {
		return nil, errors.New("nil app")
	}

	d, err := pkg.Parse(name)
	if err != nil {
		return nil, errors.Wrap(err, "parsing package name")
	}

	index, err := allLibraries(m.app)
	if err != nil {
		return nil, errors.Wrap(err, "resolving libraries")
	}

	if byVer, ok := index[d]; ok {
		// NOTE: We will use the first match for ambiguous finds
		for _, libCfg := range byVer {
			protocol, ok := registryProtocol(m.app, libCfg.Registry)
			if !ok {
				return nil, errors.Errorf("library %s references invalid registry: %s", libCfg.Name, libCfg.Registry)
			}

			return m.loadPackage(protocol, libCfg.Name, libCfg.Registry, libCfg.Version, m.InstallChecker)
		}
	}

	// If we are here, the package is not installed locally - construct a remote reference

	if d.Registry == "" {
		return nil, errors.New("cannot find package - please specify a registry")
	}

	registry, err := m.resolverFn(d.Registry)
	if err != nil {
		return nil, err
	}

	partConfig, err := registry.ResolveLibrarySpec(d.Name, d.Version)
	if err != nil {
		return nil, err
	}

	p := &remotePackage{registryName: d.Registry, partConfig: partConfig}
	return p, nil
}

type remotePackage struct {
	registryName string
	partConfig   *parts.Spec
}

var _ pkg.Package = remotePackage{}

func (p remotePackage) Name() string {
	if p.partConfig == nil {
		return ""
	}
	return p.partConfig.Name
}

func (p remotePackage) RegistryName() string {
	return p.registryName
}

func (p remotePackage) Version() string {
	if p.partConfig == nil {
		return ""
	}
	return p.partConfig.Version
}

func (p remotePackage) Description() string {
	if p.partConfig == nil {
		return ""
	}
	return p.partConfig.Description
}

func (p remotePackage) IsInstalled() (bool, error) {
	return false, nil
}

func (p remotePackage) Prototypes() (prototype.Prototypes, error) {
	return prototype.Prototypes{}, nil
}

func (p remotePackage) Path() string {
	return ""
}

func (p remotePackage) String() string {
	if p.partConfig == nil {
		return "nil"
	}
	return fmt.Sprintf("%s/%s@%s", p.registryName, p.partConfig.Name, p.partConfig.Version)
}

// Packages returns a list of Packages defined in the application.
func (m *packageManager) Packages() ([]pkg.Package, error) {
	if m.app == nil {
		return nil, errors.New("app is required")
	}

	libIndex, err := allLibraries(m.app)
	if err != nil {
		return nil, errors.Wrapf(err, "resolving libraries")
	}

	libraryConfigs := uniqueLibsByVersion(libIndex)

	packages := make([]pkg.Package, 0)

	for _, libraryConfig := range libraryConfigs {
		protocol, ok := registryProtocol(m.app, libraryConfig.Registry)
		if !ok {
			return nil, errors.Errorf("registry %q required by library %q is not defined in the configuration",
				libraryConfig.Registry, libraryConfig.Name)
		}

		p, err := m.loadPackage(protocol, libraryConfig.Name, libraryConfig.Registry, libraryConfig.Version, pkg.TrueInstallChecker{})
		if err != nil {
			return nil, err
		}

		packages = append(packages, p)
	}

	return packages, nil
}

// PackagesForEnv returns a list of Packages defined in the application, from the context
// of the specified environment.
func (m *packageManager) PackagesForEnv(e *app.EnvironmentConfig) ([]pkg.Package, error) {
	if m.app == nil {
		return nil, errors.New("nil app")
	}
	if e == nil {
		return nil, errors.New("nil environment")
	}

	globalLibs, err := m.app.Libraries()
	if err != nil {
		return nil, errors.Wrap(err, "reading libraries defined in the configuration")
	}

	packages := make([]pkg.Package, 0)

	combined := make(map[string]*app.LibraryConfig)
	// Environment-specific libraries. These take precedence.
	for k, cfg := range e.Libraries {
		combined[k] = cfg
	}
	for k, cfg := range globalLibs {
		if _, ok := combined[k]; ok {
			continue
		}
		combined[k] = cfg
	}

	for k, libraryConfig := range combined {
		protocol, ok := registryProtocol(m.app, libraryConfig.Registry)
		if !ok {
			return nil, errors.Errorf("registry %q required by library %q is not defined in the configuration",
				libraryConfig.Registry, k)
		}

		p, err := m.loadPackage(protocol, k, libraryConfig.Registry, libraryConfig.Version, pkg.TrueInstallChecker{})
		if err != nil {
			return nil, err
		}

		packages = append(packages, p)
	}

	return packages, nil
}

func (m *packageManager) RemotePackages() ([]pkg.Package, error) {
	registries, err := m.registriesFn()
	if err != nil {
		return nil, err
	}

	var pkgs []pkg.Package

	for name, r := range registries {
		spec, err := r.FetchRegistrySpec()
		if err != nil {
			return nil, err
		}

		for libName, config := range spec.Libraries {
			p := remotePackage{
				registryName: name,
				partConfig: &parts.Spec{
					Name:    libName,
					Version: config.Version,
				},
			}

			pkgs = append(pkgs, p)
		}
	}

	return pkgs, nil
}

func (m *packageManager) loadPackage(protocol Protocol, pkgName, registryName, version string, installChecker pkg.InstallChecker) (pkg.Package, error) {
	switch protocol {
	case ProtocolHelm:
		h, err := pkg.NewHelm(m.app, pkgName, registryName, version, installChecker)
		if err != nil {
			return nil, errors.Wrap(err, "loading helm package")
		}
		return h, nil
	case ProtocolFilesystem, ProtocolGitHub:
		l, err := pkg.NewLocal(m.app, pkgName, registryName, version, installChecker)
		if err != nil {
			return nil, errors.Wrapf(err, "loading %q package", protocol)
		}

		return l, nil
	default:
		return nil, errors.Errorf("package %q - registry uses unknown protocol: %q",
			fmt.Sprintf("%s/%s", registryName, pkgName), protocol)
	}
}

func (m *packageManager) Prototypes() (prototype.Prototypes, error) {
	packages, err := m.packagesFn()
	if err != nil {
		return nil, errors.Wrap(err, "loading packages")
	}

	var result prototype.Prototypes

	// Index prototypes by name
	byName := make(map[string]prototype.Prototypes)
	for _, p := range packages {
		protos, err := p.Prototypes()
		if err != nil {
			return nil, errors.Wrap(err, "loading prototypes")
		}

		for _, p := range protos {
			lst := byName[p.Name]
			lst = append(lst, p)
			byName[p.Name] = lst
		}
	}

	for _, protos := range byName {
		if len(protos) == 0 {
			continue
		}

		p := latestPrototype(protos)
		if p == nil {
			continue
		}

		result = append(result, p)
	}

	return result, nil
}

type libraryByVersion map[string]*app.LibraryConfig
type librariesByVersion map[string][]*app.LibraryConfig

// Index libraries by descriptor, each can have multiple distinct versions.
// The same library can be indexed under multiple permutations of its fully-qualified descriptor
type libraryByDesc map[pkg.Descriptor]libraryByVersion

func (index libraryByDesc) String() string {
	var sb strings.Builder

	for d, byVersion := range index {
		sb.WriteString(fmt.Sprintf("[%v]: byVersion=%v\n", d, byVersion))
	}

	return sb.String()
}

func indexLibrary(index libraryByDesc, d pkg.Descriptor, l *app.LibraryConfig) {
	byVer, ok := index[d]
	if !ok {
		byVer = libraryByVersion{}
		index[d] = byVer
	}

	// NOTE d.Version is not always equal to l.Version - we index the same
	// library under multiple descriptors to facilitate searching.
	byVer[l.Version] = l
}

func indexLibraryPermutations(index libraryByDesc, l *app.LibraryConfig) {
	if l == nil {
		return
	}

	d := pkg.Descriptor{Name: l.Name, Registry: l.Registry, Version: l.Version}
	indexLibrary(index, d, l)

	d = pkg.Descriptor{Name: l.Name, Registry: "", Version: l.Version}
	indexLibrary(index, d, l)

	d = pkg.Descriptor{Name: l.Name, Registry: l.Registry, Version: ""}
	indexLibrary(index, d, l)

	d = pkg.Descriptor{Name: l.Name, Registry: "", Version: ""}
	indexLibrary(index, d, l)
}

// Returns index of library configurations for the app.
// Libraries are indexed using multiple permutations to aid
// in search using partial keys.
func allLibraries(a app.App) (libraryByDesc, error) {
	if a == nil {
		return nil, errors.Errorf("nil app")
	}

	index := libraryByDesc{}

	libs, err := a.Libraries()
	if err != nil {
		return nil, errors.Wrapf(err, "checking libraries")
	}
	for _, l := range libs {
		indexLibraryPermutations(index, l)
	}

	envs, err := a.Environments()
	if err != nil {
		return nil, errors.Wrapf(err, "checking environments")
	}
	for _, env := range envs {
		for _, l := range env.Libraries {
			indexLibraryPermutations(index, l)
		}
	}

	return index, nil
}

// latestPrototype returns the latest prototype from the provided list.
// The list should represent different versions of the same prototype, as defined by having
// the same unqualified name. The list will not be modified.
func latestPrototype(protos prototype.Prototypes) *prototype.Prototype {
	if len(protos) == 0 {
		return nil
	}

	sorted := make(prototype.Prototypes, len(protos))
	copy(sorted, protos)
	sorted.SortByVersion()

	return sorted[len(sorted)-1]
}

// Given an index of libaries (as created by allLibraries),
// return flag list of unique libraries, as distinguished by key registry:name:version.
func uniqueLibsByVersion(libIndex libraryByDesc) []*app.LibraryConfig {
	var result = make([]*app.LibraryConfig, 0, len(libIndex))

	for d, byVersion := range libIndex {
		// Skip overly qualified indexes (remaining packages will be unique by version)
		if d.Name == "" || d.Registry == "" || d.Version != "" {
			continue
		}

		for _, l := range byVersion {
			result = append(result, l)
		}
	}

	return result
}

// IsInstalled determines whether the specified package is installed.
// Only Name is required in the descriptor, Registry and Version are inferred as "any"
// if missing. IsInstalled will make as specific a match as possible.
func (m *packageManager) IsInstalled(d pkg.Descriptor) (bool, error) {
	if m == nil {
		return false, errors.Errorf("nil receiver")
	}
	a := m.app
	if a == nil {
		return false, errors.Errorf("nil app")
	}

	if d.Name == "" {
		return false, errors.Errorf("name required")
	}

	index, err := allLibraries(a)
	if err != nil {
		return false, errors.Wrapf(err, "indexing libraries")
	}

	byVer := index[d]
	return len(byVer) > 0, nil
}

// PackageEnvironments returns a list of environments a package is installed in.
func (m *packageManager) PackageEnvironments(pkg pkg.Package) ([]*app.EnvironmentConfig, error) {
	if pkg == nil {
		return nil, errors.New("nil package")
	}

	envs, err := m.environmentsFn()
	if err != nil {
		return nil, nil
	}

	results := make([]*app.EnvironmentConfig, 0)
	for _, e := range envs {
		for _, l := range e.Libraries {
			if l.Registry == pkg.RegistryName() &&
				l.Name == pkg.Name() &&
				l.Version == pkg.Version() {
				results = append(results, e)
			}
		}
	}

	return results, nil
}
