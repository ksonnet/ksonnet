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
	"github.com/ksonnet/ksonnet/pkg/registry"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// RunRegistrySet runs `env list`
func RunRegistrySet(m map[string]interface{}) error {
	ru, err := NewRegistrySet(m)
	if err != nil {
		return err
	}

	ol := newOptionLoader(m)
	name := ol.LoadString(OptionName)
	uri := ol.LoadString(OptionURI)

	if ol.err != nil {
		return ol.err
	}

	return ru.run(name, uri)
}

type locateFn func(app.App, *app.RegistryConfig) (registry.Setter, error)

// RegistrySet lists available registries
type RegistrySet struct {
	app      app.App
	locateFn locateFn
}

// NewRegistrySet creates an instance of RegistrySet
func NewRegistrySet(m map[string]interface{}) (*RegistrySet, error) {
	ol := newOptionLoader(m)

	rs := &RegistrySet{
		app:      ol.LoadApp(),
		locateFn: defaultLocate,
	}

	if ol.err != nil {
		return nil, ol.err
	}

	return rs, nil
}

// defaultLocate passes-through to registry.Locate, but constrains the interface
// to just `registry.Setter`. The concrete type of registry.Setter is determined
// by the `spec` argument. In other words, this is a factory for registry.Setter implementations.
func defaultLocate(ksApp app.App, spec *app.RegistryConfig) (registry.Setter, error) {
	return registry.Locate(ksApp, spec)
}

// registryConfig returns a registry configuration by name from the provided App.
func registryConfig(a app.App, name string) (*app.RegistryConfig, error) {
	if name == "" {
		return nil, errors.Errorf("registry name required")
	}

	if a == nil {
		return nil, errors.Errorf("missing application")
	}

	// NOTE: app.Registries() does not currently cache app configuration
	specs, err := a.Registries()
	if err != nil {
		return nil, errors.Wrap(err, "retrieving configured registries")
	}

	cfg, ok := specs[name]
	if !ok {
		return nil, errors.Errorf("unknown registry: %v", name)
	}
	return cfg, nil
}

// run runs the registry set command.
func (rs *RegistrySet) run(name string, uri string) error {
	if rs == nil {
		return errors.Errorf("nil receiver")
	}

	cfg, err := registryConfig(rs.app, name)
	if err != nil {
		return err
	}

	return doSetURI(rs.app, rs.locateFn, cfg, uri)
}

// doSetURI sets the URI for the specified registry.
func doSetURI(a app.App, locateFn locateFn, cfg *app.RegistryConfig, uri string) error {
	if a == nil {
		return errors.Errorf("missing application")
	}

	if locateFn == nil {
		return errors.Errorf("missing registry locator function")
	}

	if uri == "" {
		return errors.Errorf("nothing to set")
	}

	// Lookup a registry.Setter implementation
	setter, err := locateFn(a, cfg)
	if err != nil || setter == nil {
		return errors.Wrap(err, "retrieving registry setter")
	}

	// Capture current settings (make a copy!!)
	var oldCfg app.RegistryConfig
	oldCfg = *setter.MakeRegistryConfig()

	log.Debugf("setting registry %v uri: %v", cfg.Name, uri)
	if err := setter.SetURI(uri); err != nil {
		return errors.Wrapf(err, "setting registry %v uri: %v", cfg.Name, uri)
	}

	// Update app only if a change has occurred
	newCfg := setter.MakeRegistryConfig()
	if oldCfg.URI == newCfg.URI {
		log.Debugf("registry %v unchanged", cfg.Name)
		return nil
	}

	// Persist changes back to app.yaml
	if err := a.UpdateRegistry(newCfg); err != nil {
		return errors.Wrapf(err, "updating registry %v in app", cfg.Name)
	}

	// Update registry cache
	if _, err := setter.FetchRegistrySpec(); err != nil {
		return errors.Wrap(err, "cache registry")
	}

	return nil
}
