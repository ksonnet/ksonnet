package metadata

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ksonnet/ksonnet/metadata/app"
	"github.com/ksonnet/ksonnet/metadata/parts"
	"github.com/ksonnet/ksonnet/metadata/registry"
	"github.com/ksonnet/ksonnet/prototype"
	str "github.com/ksonnet/ksonnet/strings"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

// AddRegistry adds a registry with `name`, `protocol`, and `uri` to
// the current ksonnet application.
func (m *manager) AddRegistry(name, protocol, uri, version string) (*registry.Spec, error) {
	appSpec, err := app.Read(m.appFS, m.rootPath)
	if err != nil {
		return nil, err
	}

	// Add registry reference to app spec.
	registryManager, err := makeGitHubRegistryManager(&app.RegistryRefSpec{
		Name:     name,
		Protocol: protocol,
		URI:      uri,
	})
	if err != nil {
		return nil, err
	}

	err = appSpec.AddRegistryRef(registryManager.RegistryRefSpec)
	if err != nil {
		return nil, err
	}

	// Retrieve the contents of registry.
	registrySpec, err := m.getOrCacheRegistry(registryManager)
	if err != nil {
		return nil, err
	}

	// Write registry specification back out to app specification.
	specBytes, err := appSpec.Marshal()
	if err != nil {
		return nil, err
	}

	err = afero.WriteFile(m.appFS, m.appYAMLPath, specBytes, defaultFilePermissions)
	if err != nil {
		return nil, err
	}

	return registrySpec, nil
}

func (m *manager) GetRegistry(name string) (*registry.Spec, string, error) {
	registryManager, protocol, err := m.getRegistryManager(name)
	if err != nil {
		return nil, "", err
	}

	regSpec, exists, err := m.registrySpecFromFile(m.registryPath(registryManager))
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
		return nil, fmt.Errorf("COuld not find registry '%s'", registryName)
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

func (m *manager) CacheDependency(registryName, libID, libName, libVersion string) (*parts.Spec, error) {
	// Retrieve application specification.
	appSpec, err := app.Read(m.appFS, m.rootPath)
	if err != nil {
		return nil, err
	}

	if _, ok := appSpec.Libraries[libName]; ok {
		return nil, fmt.Errorf("Package '%s' already exists. Use the --name flag to install this package with a unique identifier", libName)
	}

	// Retrieve registry manager for this specific registry.
	regRefSpec, exists := appSpec.GetRegistryRef(registryName)
	if !exists {
		return nil, fmt.Errorf("Registry '%s' does not exist", registryName)
	}

	registryManager, _, err := m.getRegistryManagerFor(regRefSpec)
	if err != nil {
		return nil, err
	}

	// Get all directories and files first, then write to disk. This
	// protects us from failing with a half-cached dependency because of
	// a network failure.
	directories := []string{}
	files := map[string][]byte{}
	parts, libRef, err := registryManager.ResolveLibrary(
		libID,
		libName,
		libVersion,
		func(relPath string, contents []byte) error {
			files[str.AppendToPath(m.vendorPath, relPath)] = contents
			return nil
		},
		func(relPath string) error {
			directories = append(directories, str.AppendToPath(m.vendorPath, relPath))
			return nil
		})
	if err != nil {
		return nil, err
	}

	// Add library to app specification, but wait to write it out until
	// the end, in case one of the network calls fails.
	appSpec.Libraries[libName] = libRef
	appSpecData, err := appSpec.Marshal()
	if err != nil {
		return nil, err
	}

	log.Infof("Retrieved %d files", len(files))

	for _, dir := range directories {
		if err := m.appFS.MkdirAll(dir, defaultFolderPermissions); err != nil {
			return nil, err
		}
	}

	for path, content := range files {
		if err := afero.WriteFile(m.appFS, path, content, defaultFilePermissions); err != nil {
			return nil, err
		}
	}

	err = afero.WriteFile(m.appFS, m.appYAMLPath, appSpecData, defaultFilePermissions)
	if err != nil {
		return nil, err
	}

	return parts, nil
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

func (m *manager) registryDir(regManager registry.Manager) string {
	return str.AppendToPath(m.registriesPath, regManager.RegistrySpecDir())
}

func (m *manager) registryPath(regManager registry.Manager) string {
	return str.AppendToPath(m.registriesPath, regManager.RegistrySpecFilePath())
}

func (m *manager) getRegistryManager(registryName string) (registry.Manager, string, error) {
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

func (m *manager) getRegistryManagerFor(registryRefSpec *app.RegistryRefSpec) (registry.Manager, string, error) {
	var err error
	var manager registry.Manager
	var protocol string

	switch registryRefSpec.Protocol {
	case "github":
		manager, err = makeGitHubRegistryManager(registryRefSpec)
		protocol = "github"
	default:
		return nil, "", fmt.Errorf("Invalid protocol '%s'", registryRefSpec.Protocol)
	}

	if err != nil {
		return nil, "", err
	}

	return manager, protocol, nil
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

func (m *manager) getOrCacheRegistry(gh registry.Manager) (*registry.Spec, error) {
	// Check local disk cache.
	registrySpecFile := m.registryPath(gh)
	registrySpec, exists, err := m.registrySpecFromFile(registrySpecFile)
	if !exists {
		// If failed, use the protocol to try to retrieve app specification.
		registrySpec, err = gh.FetchRegistrySpec()
		if err != nil {
			return nil, err
		}

		registrySpecBytes, err := registrySpec.Marshal()
		if err != nil {
			return nil, err
		}

		// NOTE: We call mkdir after getting the registry spec, since a
		// network call might fail and leave this half-initialized empty
		// directory.
		registrySpecDir := str.AppendToPath(m.registriesPath, gh.RegistrySpecDir())
		err = m.appFS.MkdirAll(registrySpecDir, defaultFolderPermissions)
		if err != nil {
			return nil, err
		}

		err = afero.WriteFile(m.appFS, registrySpecFile, registrySpecBytes, defaultFilePermissions)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	return registrySpec, nil
}
