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

package app

import (
	"path/filepath"

	"github.com/spf13/afero"
)

type baseApp struct {
	root string
	fs   afero.Fs
}

func newBaseApp(fs afero.Fs, root string) *baseApp {
	return &baseApp{
		fs:   fs,
		root: root,
	}
}

func (ba *baseApp) save(spec *Spec) error {
	return write(ba.fs, ba.root, spec)
}

func (ba *baseApp) load() (*Spec, error) {
	spec, err := read(ba.fs, ba.root)
	if err != nil {
		return nil, err
	}

	return spec, nil
}

func (ba *baseApp) AddRegistry(newReg *RegistryRefSpec, isOverride bool) error {
	spec, err := ba.load()
	if err != nil {
		return err
	}

	if newReg.Name == "" {
		return ErrRegistryNameInvalid
	}

	_, exists := spec.Registries[newReg.Name]
	if exists && !isOverride {
		return ErrRegistryExists
	}

	newReg.isOverride = isOverride
	spec.Registries[newReg.Name] = newReg

	return ba.save(spec)
}

func (ba *baseApp) UpdateLib(name string, libSpec *LibraryRefSpec) error {
	spec, err := ba.load()
	if err != nil {
		return err
	}

	spec.Libraries[name] = libSpec

	return ba.save(spec)
}

func (ba *baseApp) Fs() afero.Fs {
	return ba.fs
}

func (ba *baseApp) Root() string {
	return ba.root
}

func (ba *baseApp) EnvironmentParams(envName string) (string, error) {
	envParamsPath := filepath.Join(ba.Root(), EnvironmentDirName, envName, "params.libsonnet")
	b, err := afero.ReadFile(ba.Fs(), envParamsPath)
	if err != nil {
		return "", err
	}

	return string(b), nil
}
