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
	"path/filepath"

	"github.com/ksonnet/ksonnet/metadata/app"
	"github.com/pkg/errors"
)

var (
	// DefaultManager is the default manager for registries.
	DefaultManager = &defaultManager{}
)

// Locate locates a registry given a spec.
func Locate(a app.App, spec *app.RegistryRefSpec) (Registry, error) {
	switch spec.Protocol {
	case ProtocolGitHub:
		return NewGitHub(spec)
	case ProtocolFilesystem:
		return NewFs(a, spec)
	default:
		return nil, errors.Errorf("invalid registry protocol %q", spec.Protocol)
	}
}

// TODO: add this to App
func root(a app.App) string {
	return filepath.Join(a.Root(), ".ksonnet", "registries")
}

func makePath(a app.App, r Registry) string {
	path := r.RegistrySpecFilePath()
	if filepath.IsAbs(path) {
		return path
	}

	return filepath.Join(root(a), path)
}

// Manager is a manager for registry related actions.
type Manager interface {
	Add(a app.App, name, protoocol, uri, version string) (*Spec, error)
	// Registries returns a list of alphabetically sorted registries. The
	// registries are sorted by name.
	List(ksApp app.App) ([]Registry, error)
}

type defaultManager struct{}

var _ Manager = (*defaultManager)(nil)

func (dm *defaultManager) List(ksApp app.App) ([]Registry, error) {

	var registries []Registry
	appRegistries, err := ksApp.Registries()
	if err != nil {
		return nil, err
	}
	for name, regRef := range appRegistries {
		regRef.Name = name
		r, err := Locate(ksApp, regRef)
		if err != nil {
			return nil, err
		}
		registries = append(registries, r)
	}

	return registries, nil
}

func (dm *defaultManager) Add(a app.App, name, protoocol, uri, version string) (*Spec, error) {
	return Add(a, name, protoocol, uri, version)
}
