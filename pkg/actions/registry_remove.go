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
)

// RunRegistryRemove runs `registry remove`
func RunRegistryRemove(m map[string]interface{}) error {
	rr, err := NewRegistryRemove(m)
	if err != nil {
		return err
	}

	return rr.Run()
}

// RegistryRemove removes a registry.
type RegistryRemove struct {
	app  app.App
	name string

	registryRemoveFn func(a app.App, name string) error
}

// NewRegistryRemove creates an instance of RegistryRemove.
func NewRegistryRemove(m map[string]interface{}) (*RegistryRemove, error) {
	ol := newOptionLoader(m)

	rr := &RegistryRemove{
		app:  ol.LoadApp(),
		name: ol.LoadString(OptionName),

		registryRemoveFn: registry.Remove,
	}

	if ol.err != nil {
		return nil, ol.err
	}

	return rr, nil
}

// Run removes a registry.
func (rr *RegistryRemove) Run() error {
	err := rr.registryRemoveFn(rr.app, rr.name)
	return err
}
