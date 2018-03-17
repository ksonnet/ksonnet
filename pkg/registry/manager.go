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

import "github.com/ksonnet/ksonnet/metadata/app"

var (
	// DefaultManager is the default manager for registries.
	DefaultManager = &defaultManager{}
)

// Registry is a registry.
type Registry interface {
	Name() string
	Protocol() string
	URI() string
}

// Manager is a manager for registry related actions.
type Manager interface {
	// Registries returns a list of alphabetically sorted registries. The
	// registries are sorted by name.
	Registries(ksApp app.App) ([]Registry, error)
}

type defaultManager struct{}

var _ Manager = (*defaultManager)(nil)

func (dm *defaultManager) Registries(ksApp app.App) ([]Registry, error) {

	var registries []Registry
	for name, regRef := range ksApp.Registries() {
		registries = append(registries, NewGitHub(name, regRef))
	}

	return registries, nil
}
