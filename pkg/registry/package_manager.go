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

	// Prototypes lists prototypes.
	Prototypes() (prototype.Prototypes, error)

	InstalledChecker
}

// packageManager is an implementation of PackageManager.
type packageManager struct {
	app app.App

	InstallChecker pkg.InstallChecker
	packagesFn     func() ([]pkg.Package, error)
}

var _ PackageManager = (*packageManager)(nil)

// NewPackageManager creates an instance of PackageManager.
func NewPackageManager(a app.App) PackageManager {
	pm := packageManager{
		app:            a,
		InstallChecker: &pkg.DefaultInstallChecker{App: a},
	}
	pm.packagesFn = pm.Packages

	return &pm
}

// Find finds a package by name. Package names have the format `<registry>/<library>@<version>`.
// Remote registries may be consulted if the package is not installed locally.
func (m *packageManager) Find(name string) (pkg.Package, error) {
	d, err := pkg.Parse(name)
	if err != nil {
		return nil, errors.Wrap(err, "parsing package name")
	}

	if d.Registry == "" {
		packages, err := m.Packages()
		if err != nil {
			return nil, errors.Wrap(err, "loading packages")
		}
		for _, p := range packages {
			if p.Name() == name {
				return p, nil
			}
		}

		return nil, errors.Errorf("package %q was not found", name)
	}

	registryConfigs, err := m.app.Registries()
	if err != nil {
		return nil, errors.Wrap(err, "loading registry configurations")
	}

	registryConfig, ok := registryConfigs[d.Registry]
	if !ok {
		return nil, errors.Errorf("registry %q not found", d.Registry)
	}

	registry, err := Locate(m.app, registryConfig)
	if err != nil {
		return nil, err
	}

	libraryConfigs, err := m.app.Libraries()
	if err != nil {
		return nil, errors.Wrap(err, "reading libraries defined in the configuration")
	}

	libraryConfig, ok := libraryConfigs[d.Name]
	if ok {
		return m.loadPackage(registry.MakeRegistryConfig(), d.Name, d.Registry, libraryConfig.Version)
	}

	// TODO - Check libraries configured under environments

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

var _ pkg.Package = (*remotePackage)(nil)

func (p *remotePackage) Name() string {
	if p == nil || p.partConfig == nil {
		return ""
	}
	return p.partConfig.Name
}

func (p *remotePackage) RegistryName() string {
	if p == nil {
		return ""
	}
	return p.registryName
}

func (p *remotePackage) Version() string {
	if p == nil || p.partConfig == nil {
		return ""
	}
	return p.partConfig.Version
}

func (p *remotePackage) Description() string {
	if p == nil || p.partConfig == nil {
		return ""
	}
	return p.partConfig.Description
}

func (p *remotePackage) IsInstalled() (bool, error) {
	return false, nil
}

func (p *remotePackage) Prototypes() (prototype.Prototypes, error) {
	return prototype.Prototypes{}, nil
}

func (p *remotePackage) Path() string {
	return ""
}

func (p *remotePackage) String() string {
	if p == nil || p.partConfig == nil {
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

	registryConfigs, err := m.app.Registries()
	if err != nil {
		return nil, errors.Wrap(err, "reading registries defined in the configuration")
	}

	packages := make([]pkg.Package, 0)

	for _, libraryConfig := range libraryConfigs {
		registryConfig, ok := registryConfigs[libraryConfig.Registry]
		if !ok {
			return nil, errors.Errorf("registry %q required by library %q is not defined in the configuration",
				libraryConfig.Registry, libraryConfig.Name)
		}

		p, err := m.loadPackage(registryConfig, libraryConfig.Name, libraryConfig.Registry, libraryConfig.Version)
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

	registryConfigs, err := m.app.Registries()
	if err != nil {
		return nil, errors.Wrap(err, "reading registries defined in the configuration")
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
		registryConfig, ok := registryConfigs[libraryConfig.Registry]
		if !ok {
			return nil, errors.Errorf("registry %q required by library %q is not defined in the configuration",
				libraryConfig.Registry, k)
		}

		p, err := m.loadPackage(registryConfig, k, libraryConfig.Registry, libraryConfig.Version)
		if err != nil {
			return nil, err
		}

		packages = append(packages, p)
	}

	return packages, nil
}

func (m *packageManager) loadPackage(registryConfig *app.RegistryConfig, pkgName, registryName, version string) (pkg.Package, error) {
	switch protocol := registryConfig.Protocol; Protocol(protocol) {
	case ProtocolHelm:
		h, err := pkg.NewHelm(m.app, pkgName, registryName, version, m.InstallChecker)
		if err != nil {
			return nil, errors.Wrap(err, "loading helm package")
		}
		return h, nil
	case ProtocolFilesystem, ProtocolGitHub:
		l, err := pkg.NewLocal(m.app, pkgName, registryName, version, m.InstallChecker)
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
