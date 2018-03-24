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
	"github.com/davecgh/go-spew/spew"
	"github.com/ksonnet/ksonnet/metadata/app"
	"github.com/ksonnet/ksonnet/pkg/pkg"
	"github.com/ksonnet/ksonnet/pkg/registry"
)

// DepCacher is a function that caches a dependency.j
type DepCacher func(app.App, pkg.Descriptor, string) error

// PkgInstallDepCacher sets the dep cacher for pkg install.
func PkgInstallDepCacher(dc DepCacher) PkgInstallOpt {
	return func(pi *PkgInstall) {
		pi.depCacher = dc
	}
}

// PkgInstallOpt is an option for configuring PkgInstall.
type PkgInstallOpt func(*PkgInstall)

// RunPkgInstall runs `pkg install`
func RunPkgInstall(ksApp app.App, libName, customName string, opts ...PkgInstallOpt) error {
	pi, err := NewPkgInstall(ksApp, libName, customName, opts...)
	if err != nil {
		return err
	}

	return pi.Run()
}

// PkgInstall lists namespaces.
type PkgInstall struct {
	app        app.App
	libName    string
	customName string
	depCacher  DepCacher
}

// NewPkgInstall creates an instance of PkgInstall.
func NewPkgInstall(ksApp app.App, libName, name string, opts ...PkgInstallOpt) (*PkgInstall, error) {
	nl := &PkgInstall{
		app:        ksApp,
		libName:    libName,
		customName: name,
		depCacher:  registry.CacheDependency,
	}

	for _, opt := range opts {
		opt(nl)
	}

	return nl, nil
}

// Run lists namespaces.
func (pi *PkgInstall) Run() error {
	d, customName, err := pi.parseDepSpec()
	if err != nil {
		return err
	}

	spew.Dump(d)

	return pi.depCacher(pi.app, d, customName)
}

func (pi *PkgInstall) parseDepSpec() (pkg.Descriptor, string, error) {
	d, err := pkg.ParseName(pi.libName)
	if err != nil {
		return pkg.Descriptor{}, "", err
	}

	customName := pi.customName
	if customName == "" {
		customName = d.Part
	}

	return d, customName, nil
}
