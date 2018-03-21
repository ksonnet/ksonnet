// Copyright 2018 The kubecfg authors
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

package metadata

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ksonnet/ksonnet/metadata/app"
	"github.com/ksonnet/ksonnet/pkg/parts"
	"github.com/ksonnet/ksonnet/pkg/registry"
	"github.com/ksonnet/ksonnet/prototype"
	str "github.com/ksonnet/ksonnet/strings"
	"github.com/spf13/afero"
)

func (m *manager) GetRegistry(name string) (*registry.Spec, string, error) {
	r, protocol, err := m.getRegistryManager(name)
	if err != nil {
		return nil, "", err
	}

	regSpec, exists, err := m.registrySpecFromFile(m.registryPath(r))
	if !exists {
		return nil, "", fmt.Errorf("Registry '%s' does not exist", name)
	} else if err != nil {
		return nil, "", err
	}

	return regSpec, protocol, nil
}

func (m *manager) GetPackage(registryName, libID string) (*parts.Spec, error) {
	// Retrieve application specification.
	appSpec, err := app.Read(m.appFS, m.rootPath)
	if err != nil {
		return nil, err
	}

	regRefSpec, ok := appSpec.GetRegistryRef(registryName)
	if !ok {
		return nil, fmt.Errorf("Could not find registry '%s'", registryName)
	}

	registryManager, _, err := m.getRegistryManagerFor(regRefSpec)
	if err != nil {
		return nil, err
	}

	partsSpec, err := registryManager.ResolveLibrarySpec(libID, regRefSpec.GitVersion.CommitSHA)
	if err != nil {
		return nil, err
	}

	protoSpecs, err := m.GetPrototypesForDependency(registryName, libID)
	if err != nil {
		return nil, err
	}

	for _, protoSpec := range protoSpecs {
		partsSpec.Prototypes = append(partsSpec.Prototypes, protoSpec.Name)
	}

	return partsSpec, nil
}

func (m *manager) GetDependency(libName string) (*parts.Spec, error) {
	// Retrieve application specification.
	appSpec, err := app.Read(m.appFS, m.rootPath)
	if err != nil {
		return nil, err
	}

	libRef, ok := appSpec.Libraries[libName]
	if !ok {
		return nil, fmt.Errorf("Library '%s' is not a dependency in current ksonnet app", libName)
	}

	partsYAMLPath := str.AppendToPath(m.vendorPath, libRef.Registry, libName, partsYAMLFile)
	partsBytes, err := afero.ReadFile(m.appFS, partsYAMLPath)
	if err != nil {
		return nil, err
	}

	partsSpec, err := parts.Unmarshal(partsBytes)
	if err != nil {
		return nil, err
	}

	protoSpecs, err := m.GetPrototypesForDependency(libRef.Registry, libName)
	if err != nil {
		return nil, err
	}

	for _, protoSpec := range protoSpecs {
		partsSpec.Prototypes = append(partsSpec.Prototypes, protoSpec.Name)
	}

	return partsSpec, nil
}

func (m *manager) GetPrototypesForDependency(registryName, libID string) (prototype.SpecificationSchemas, error) {
	// TODO: Remove `registryName` when we flatten vendor/.
	specs := prototype.SpecificationSchemas{}
	protos := str.AppendToPath(m.vendorPath, registryName, libID, "prototypes")
	exists, err := afero.DirExists(m.appFS, protos)
	if err != nil {
		return nil, err
	} else if !exists {
		return prototype.SpecificationSchemas{}, nil // No prototypes to report.
	}

	err = afero.Walk(
		m.appFS,
		protos,
		func(path string, info os.FileInfo, err error) error {
			if info.IsDir() || filepath.Ext(path) != ".jsonnet" {
				return nil
			}

			protoJsonnet, err := afero.ReadFile(m.appFS, path)
			if err != nil {
				return err
			}

			protoSpec, err := prototype.FromJsonnet(string(protoJsonnet))
			if err != nil {
				return err
			}
			specs = append(specs, protoSpec)
			return nil
		})
	if err != nil {
		return nil, err
	}

	return specs, nil
}

func (m *manager) GetAllPrototypes() (prototype.SpecificationSchemas, error) {
	appSpec, err := app.Read(m.appFS, m.rootPath)
	if err != nil {
		return nil, err
	}

	specs := prototype.SpecificationSchemas{}
	for _, lib := range appSpec.Libraries {
		depProtos, err := m.GetPrototypesForDependency(lib.Registry, lib.Name)
		if err != nil {
			return nil, err
		}
		specs = append(specs, depProtos...)
	}

	return specs, nil
}

func (m *manager) registryDir(regManager registry.Registry) string {
	return str.AppendToPath(m.registriesPath, regManager.RegistrySpecDir())
}

func (m *manager) registryPath(regManager registry.Registry) string {
	path := regManager.RegistrySpecFilePath()
	if filepath.IsAbs(path) {
		return path
	}
	return str.AppendToPath(m.registriesPath, regManager.RegistrySpecFilePath())
}

func (m *manager) getRegistryManager(registryName string) (registry.Registry, string, error) {
	appSpec, err := app.Read(m.appFS, m.rootPath)
	if err != nil {
		return nil, "", err
	}

	regRefSpec, exists := appSpec.GetRegistryRef(registryName)
	if !exists {
		return nil, "", fmt.Errorf("Registry '%s' does not exist", registryName)
	}

	return m.getRegistryManagerFor(regRefSpec)
}

func (m *manager) getRegistryManagerFor(registryRefSpec *app.RegistryRefSpec) (registry.Registry, string, error) {
	a, err := m.App()
	if err != nil {
		return nil, "", err
	}

	r, err := registry.Locate(a, registryRefSpec)
	if err != nil {
		return nil, "", err
	}

	return r, r.Protocol(), nil
}

func (m *manager) registrySpecFromFile(path string) (*registry.Spec, bool, error) {
	exists, err := afero.Exists(m.appFS, path)
	if err != nil {
		return nil, false, err
	}

	isDir, err := afero.IsDir(m.appFS, path)
	if err != nil {
		return nil, false, err
	}

	// NOTE: case where directory of the same name exists should be
	// fine, most filesystems allow you to have a directory and file of
	// the same name.
	if exists && !isDir {
		registrySpecBytes, err := afero.ReadFile(m.appFS, path)
		if err != nil {
			return nil, false, err
		}

		registrySpec, err := registry.Unmarshal(registrySpecBytes)
		if err != nil {
			return nil, false, err
		}
		return registrySpec, true, nil
	}

	return nil, false, nil
}
