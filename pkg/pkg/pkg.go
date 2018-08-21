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

package pkg

import (
	"fmt"

	"github.com/ksonnet/ksonnet/pkg/app"
	"github.com/ksonnet/ksonnet/pkg/prototype"
	"github.com/pkg/errors"
)

const (
	partsYAML = "parts.yaml"
)

type pkg struct {
	a              app.App
	name           string
	registryName   string
	version        string
	installChecker InstallChecker
}

// Name returns the name for the package.
func (p *pkg) Name() string {
	if p == nil {
		return ""
	}
	return p.name
}

// RegistryName returns the registry name for the package.
func (p *pkg) RegistryName() string {
	if p == nil {
		return ""
	}
	return p.registryName
}

// Version returns the package version, or empty string if the package is unversioned.
func (p *pkg) Version() string {
	if p == nil {
		return ""
	}
	return p.version
}

// IsInstalled returns true if the package is installed.
func (p *pkg) IsInstalled() (bool, error) {
	if p == nil {
		return false, errors.Errorf("nil receiver")
	}
	if p.installChecker == nil {
		return false, errors.Errorf("nil installChecker")
	}
	return p.installChecker.IsInstalled(p.name)
}

// String implements Stringer
func (p *pkg) String() string {
	if p == nil {
		return "nil"
	}

	// TODO exclude @ if verion is empty
	return fmt.Sprintf("%v/%v@%v", p.registryName, p.name, p.version)
}

// InstallChecker checks if a package is installed.
type InstallChecker interface {
	// IsInstalled returns true if a package is installed.
	IsInstalled(name string) (bool, error)
}

// DefaultInstallChecker checks if a package is installed.
type DefaultInstallChecker struct {
	App app.App
}

// IsInstalled returns true if the package is installed. a package is installed if it
// has a libraries entry in app.yaml (globally or under an environment)
func (ic *DefaultInstallChecker) IsInstalled(name string) (bool, error) {
	if ic.App == nil {
		return false, errors.New("app is nil")
	}

	d, err := Parse(name)
	if err != nil {
		return false, errors.Wrapf(err, "parsing package descriptor: %s", name)
	}

	libs, err := ic.App.Libraries()
	if err != nil {
		return false, errors.Wrapf(err, "checking if package %q is installed", name)
	}

	var isGlobal bool
	if l, ok := libs[d.Name]; ok {
		if d.Version == "" || l.Version == d.Version {
			isGlobal = true
		}
	}

	envs, err := ic.App.Environments()
	if err != nil {
		return false, errors.Wrapf(err, "checking for package %q references in environments", name)
	}

	var isLocal bool
	for _, e := range envs {
		if l, ok := e.Libraries[d.Name]; ok {
			if d.Version == "" || l.Version == d.Version {
				isLocal = true
				break
			}
		}
	}

	isInstalled := isGlobal || isLocal
	return isInstalled, nil
}

// TrueInstallChecker implements an always-true InstallChecker.
type TrueInstallChecker struct{}

// IsInstalled always returns true, signaling we knew the package was installed when it was
// bound to this installChecker.
func (ic TrueInstallChecker) IsInstalled(name string) (bool, error) {
	return true, nil
}

// Package is a ksonnet package.
type Package interface {
	// Name returns the name of the package.
	Name() string

	// RegistryName returns the registry name of the package.
	RegistryName() string

	// Version returns the package version, or empty string if the package is unversioned.
	Version() string

	// IsInstalled returns true if the package is installed.
	IsInstalled() (bool, error)

	// Description retrurns the package description
	Description() string

	// Prototypes returns prototypes defined in the package.
	Prototypes() (prototype.Prototypes, error)

	// Path returns local directory for vendoring the package.
	Path() string

	fmt.Stringer
}
