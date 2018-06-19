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
	"github.com/ksonnet/ksonnet/pkg/app"
	"github.com/ksonnet/ksonnet/pkg/parts"
	"github.com/ksonnet/ksonnet/pkg/pkg"
	"github.com/ksonnet/ksonnet/pkg/prototype"
	"github.com/pkg/errors"
)

// PackageManager is a package manager.
type PackageManager interface {
	Find(string) (pkg.Package, error)

	// Packages lists packages.
	Packages() ([]pkg.Package, error)

	// Prototypes lists prototypes.
	Prototypes() (prototype.Prototypes, error)
}

// packageManager is an implementation of PackageManager.
type packageManager struct {
	app app.App

	InstallChecker pkg.InstallChecker
}

var _ PackageManager = (*packageManager)(nil)

// NewPackageManager creates an instance of PackageManager.
func NewPackageManager(a app.App) PackageManager {
	return &packageManager{
		app:            a,
		InstallChecker: &pkg.DefaultInstallChecker{App: a},
	}
}

// Find finds a package by name. Package names have the format `<registry>/<library>@<version>`.
func (m *packageManager) Find(name string) (pkg.Package, error) {
	d, err := pkg.ParseName(name)
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

	libraryConfig, ok := libraryConfigs[d.Part]
	if ok {
		return m.loadPackage(registry.MakeRegistryRefSpec(), d.Part, d.Registry, libraryConfig.Version)
	}

	partConfig, err := registry.ResolveLibrarySpec(d.Part, d.Version)
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
	return p.partConfig.Name
}

func (p *remotePackage) RegistryName() string {
	return p.registryName
}

func (p *remotePackage) Description() string {
	return p.partConfig.Description
}

func (p *remotePackage) IsInstalled() (bool, error) {
	return false, nil
}

func (p *remotePackage) Prototypes() (prototype.Prototypes, error) {
	return prototype.Prototypes{}, nil
}

// Packages returns a list of Packages defined in the application.
func (m *packageManager) Packages() ([]pkg.Package, error) {
	if m.app == nil {
		return nil, errors.New("app is required")
	}

	libraryConfigs, err := m.app.Libraries()
	if err != nil {
		return nil, errors.Wrap(err, "reading libraries defined in the configuration")
	}

	registryConfigs, err := m.app.Registries()
	if err != nil {
		return nil, errors.Wrap(err, "reading registries defined in the configuration")
	}

	packages := make([]pkg.Package, 0)

	for k, libraryConfig := range libraryConfigs {
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

func (m *packageManager) loadPackage(registryConfig *app.RegistryRefSpec, pkgName, registryName, version string) (pkg.Package, error) {
	switch protocol := registryConfig.Protocol; Protocol(protocol) {
	case ProtocolHelm:
		h, err := pkg.NewHelm(m.app, pkgName, registryName, version, m.InstallChecker)
		if err != nil {
			return nil, errors.Wrap(err, "loading helm package")
		}
		return h, nil
	case ProtocolFilesystem, ProtocolGitHub:
		l, err := pkg.NewLocal(m.app, pkgName, registryName, m.InstallChecker)
		if err != nil {
			return nil, errors.Wrapf(err, "loading %q package", protocol)
		}

		return l, nil
	default:
		return nil, errors.Errorf("library %q has a reference to unknown prototypes %q",
			pkgName, protocol)
	}
}

func (m *packageManager) Prototypes() (prototype.Prototypes, error) {
	packages, err := m.Packages()
	if err != nil {
		return nil, errors.Wrap(err, "loading packages")
	}

	var prototypes prototype.Prototypes

	for _, p := range packages {
		protos, err := p.Prototypes()
		if err != nil {
			return nil, errors.Wrap(err, "loading prototypes")
		}

		prototypes = append(prototypes, protos...)
	}

	return prototypes, nil
}
