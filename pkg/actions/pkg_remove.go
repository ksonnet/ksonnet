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
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// PkgRemove removes packages
type PkgRemove struct {
	app         app.App
	pkgName     string
	envName     string
	checker     registry.InstalledChecker
	gc          registry.GarbageCollector
	libUpdateFn libUpdater
}

// NewPkgRemove creates an instance of PkgInstall
func NewPkgRemove(m map[string]interface{}) (*PkgRemove, error) {
	ol := newOptionLoader(m)

	a := ol.LoadApp()
	if ol.err != nil {
		return nil, ol.err
	}

	pm := registry.NewPackageManager(a)

	pr := &PkgRemove{
		app:         a,
		pkgName:     ol.LoadString(OptionPkgName),
		envName:     ol.LoadOptionalString(OptionEnvName),
		libUpdateFn: a.UpdateLib,
		gc:          registry.NewGarbageCollector(a.Fs(), pm, a.VendorPath()),
	}

	if ol.err != nil {
		return nil, ol.err
	}

	return pr, nil
}

// RunPkgRemove runs `pkg list`
func RunPkgRemove(m map[string]interface{}) error {
	pr, err := NewPkgRemove(m)
	if err != nil {
		return err
	}

	return pr.Run()
}

// Run removes packages
func (pr *PkgRemove) Run() error {
	desc, err := pkg.Parse(pr.pkgName)
	if err != nil {
		return err
	}

	oldCfg, err := pr.libUpdateFn(desc.Name, pr.envName, nil)
	if err != nil {
		return err
	}

	if oldCfg == nil {
		return nil
	}

	log.Infof("Removing package %v", oldCfg)

	// Optionally remove any orphaned vendor directories
	if err := pr.gc.RemoveOrphans(pkg.Descriptor{
		Registry: oldCfg.Registry,
		Name:     oldCfg.Name,
		Version:  oldCfg.Version,
	}); err != nil {
		return errors.Wrapf(err, "garbage collection for package %v", oldCfg)
	}

	return nil
}
