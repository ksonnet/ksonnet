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
	"github.com/ksonnet/ksonnet/pkg/app"
	"github.com/ksonnet/ksonnet/pkg/prototype"
	"github.com/pkg/errors"
)

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
// has a libraries entry in app.yaml.
func (ic *DefaultInstallChecker) IsInstalled(name string) (bool, error) {
	if ic.App == nil {
		return false, errors.New("app is nil")
	}

	libs, err := ic.App.Libraries()
	if err != nil {
		return false, errors.Wrapf(err, "checking if package %q is installed", name)
	}

	_, isInstalled := libs[name]
	return isInstalled, nil
}

// Package is a ksonnet package.
type Package interface {
	// Name returns the name of the package.
	Name() string

	// RegistryName returns the registry name of the package.
	RegistryName() string

	// IsInstalled returns true if the package is installed.
	IsInstalled() (bool, error)

	// Description retrurns the package description
	Description() string

	// Prototypes returns prototypes defined in the package.
	Prototypes() (prototype.Prototypes, error)
}
