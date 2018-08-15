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
	"github.com/ksonnet/ksonnet/pkg/app"
	"github.com/ksonnet/ksonnet/pkg/pkg"
	"github.com/ksonnet/ksonnet/pkg/registry"
)

type libCacher func(app.App, registry.InstalledChecker, pkg.Descriptor, string, bool) (*app.LibraryConfig, error)

type libUpdater func(name string, env string, spec *app.LibraryConfig) error

// RunPkgInstall runs `pkg install`
func RunPkgInstall(m map[string]interface{}) error {
	pi, err := NewPkgInstall(m)
	if err != nil {
		return err
	}

	return pi.Run()
}

// PkgInstall installs packages.
type PkgInstall struct {
	app         app.App
	libName     string
	customName  string
	envName     string
	force       bool
	checker     registry.InstalledChecker
	libCacherFn libCacher
	libUpdateFn libUpdater
}

// NewPkgInstall creates an instance of PkgInstall.
func NewPkgInstall(m map[string]interface{}) (*PkgInstall, error) {
	ol := newOptionLoader(m)

	a := ol.LoadApp()
	if ol.err != nil {
		return nil, ol.err
	}
	httpClient := ol.LoadHTTPClient()
	httpClientOpt := registry.HTTPClientOpt(httpClient)

	nl := &PkgInstall{
		app:        a,
		libName:    ol.LoadString(OptionLibName),
		customName: ol.LoadString(OptionName),
		force:      ol.LoadBool(OptionForce),
		envName:    ol.LoadOptionalString(OptionEnvName),
		checker:    registry.NewPackageManager(a, httpClientOpt),

		libCacherFn: func(a app.App, checker registry.InstalledChecker, d pkg.Descriptor, customName string, force bool) (*app.LibraryConfig, error) {
			return registry.CacheDependency(a, checker, d, customName, force, httpClient)
		},
		libUpdateFn: a.UpdateLib,
	}

	if ol.err != nil {
		return nil, ol.err
	}

	return nl, nil
}

// Run installs packages.
func (pi *PkgInstall) Run() error {
	d, customName, err := pi.parseDepSpec()
	if err != nil {
		return err
	}

	libCfg, err := pi.libCacherFn(pi.app, pi.checker, d, customName, pi.force)
	if err != nil {
		return err
	}

	err = pi.libUpdateFn(d.Name, pi.envName, libCfg)
	if err != nil {
		return err
	}

	return nil
}

func (pi *PkgInstall) parseDepSpec() (pkg.Descriptor, string, error) {
	d, err := pkg.Parse(pi.libName)
	if err != nil {
		return pkg.Descriptor{}, "", err
	}

	customName := pi.customName
	if customName == "" {
		customName = d.Name
	}

	return d, customName, nil
}
