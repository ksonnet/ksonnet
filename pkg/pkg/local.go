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
	"os"
	"path/filepath"

	"github.com/ksonnet/ksonnet/pkg/parts"

	"github.com/ksonnet/ksonnet/pkg/app"
	"github.com/ksonnet/ksonnet/pkg/prototype"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
)

// Local is a package based on vendored contents.
type Local struct {
	a              app.App
	name           string
	registryName   string
	config         *parts.Spec
	installChecker InstallChecker
}

var _ Package = (*Local)(nil)

// NewLocal creates an instance of Local.
func NewLocal(a app.App, name, registryName string, installChecker InstallChecker) (*Local, error) {
	if installChecker == nil {
		installChecker = &DefaultInstallChecker{App: a}
	}

	partsPath := filepath.Join(a.VendorPath(), registryName, name, "parts.yaml")
	b, err := afero.ReadFile(a.Fs(), partsPath)
	if err != nil {
		return nil, errors.Wrap(err, "reading package configuration")
	}

	config, err := parts.Unmarshal(b)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshalling package configuration")
	}

	return &Local{
		a:              a,
		name:           name,
		registryName:   registryName,
		config:         config,
		installChecker: installChecker,
	}, nil
}

// Name returns the name for the package.
func (l *Local) Name() string {
	return l.name
}

// RegistryName returns the registry name for the package.
func (l *Local) RegistryName() string {
	return l.registryName
}

// IsInstalled returns true if the package is installed.
func (l *Local) IsInstalled() (bool, error) {
	return l.installChecker.IsInstalled(l.Name())
}

// Description returns the description for the package. The description
// is ready from the package's parts.yaml.
func (l *Local) Description() string {
	return l.config.Description
}

// Prototypes returns prototypes for this package. Prototypes are defined in the
// package's `prototypes` directory.
func (l *Local) Prototypes() (prototype.Prototypes, error) {
	var prototypes prototype.Prototypes

	protoPath := filepath.Join(l.a.VendorPath(), l.registryName, l.name, "prototypes")
	exists, err := afero.DirExists(l.a.Fs(), protoPath)
	if err != nil {
		return nil, err
	}

	if !exists {
		return prototypes, nil
	}

	err = afero.Walk(l.a.Fs(), protoPath, func(path string, fi os.FileInfo, err error) error {
		if fi.IsDir() || filepath.Ext(path) != ".jsonnet" {
			return nil
		}

		data, err := afero.ReadFile(l.a.Fs(), path)
		if err != nil {
			return err
		}

		spec, err := prototype.DefaultBuilder(string(data))
		if err != nil {
			return err
		}

		prototypes = append(prototypes, spec)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return prototypes, nil
}
