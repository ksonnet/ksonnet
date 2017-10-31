package metadata

import (
	"encoding/json"
	"fmt"

	"github.com/ksonnet/ksonnet/metadata/app"
	"github.com/ksonnet/ksonnet/metadata/registry"
	"github.com/spf13/afero"
)

// AddRegistry adds a registry with `name`, `protocol`, and `uri` to
// the current ksonnet application.
func (m *manager) AddRegistry(name, protocol, uri string) (*registry.Spec, error) {
	app, err := m.AppSpec()
	if err != nil {
		return nil, err
	}

	// Retrieve or create registry specification.
	registryRef, err := app.AddRegistryRef(name, protocol, uri)
	if err != nil {
		return nil, err
	}

	// Retrieve the contents of registry.
	registrySpec, err := m.getOrCacheRegistry(registryRef)
	if err != nil {
		return nil, err
	}

	// Write registry specification back out to app specification.
	specBytes, err := app.Marshal()
	if err != nil {
		return nil, err
	}

	err = afero.WriteFile(m.appFS, string(m.appYAMLPath), specBytes, defaultFilePermissions)
	if err != nil {
		return nil, err
	}

	return registrySpec, nil
}

func (m *manager) registryDir(regManager registry.Manager) AbsPath {
	return appendToAbsPath(m.registriesPath, regManager.VersionsDir())
}

func (m *manager) registryPath(regManager registry.Manager) AbsPath {
	return appendToAbsPath(m.registriesPath, regManager.SpecPath())
}

func (m *manager) getOrCacheRegistry(registryRefSpec *app.RegistryRefSpec) (*registry.Spec, error) {
	switch registryRefSpec.Protocol {
	case "github":
		break
	default:
		return nil, fmt.Errorf("Invalid protocol '%s'", registryRefSpec.Protocol)
	}

	// Check local disk cache.
	gh, err := makeGitHubRegistryManager(registryRefSpec)
	if err != nil {
		return nil, err
	}
	registrySpecFile := string(m.registryPath(gh))

	exists, _ := afero.Exists(m.appFS, registrySpecFile)
	isDir, _ := afero.IsDir(m.appFS, registrySpecFile)
	if exists && !isDir {
		registrySpecBytes, err := afero.ReadFile(m.appFS, registrySpecFile)
		if err != nil {
			return nil, err
		}

		registrySpec := registry.Spec{}
		err = json.Unmarshal(registrySpecBytes, &registrySpec)
		if err != nil {
			return nil, err
		}
		return &registrySpec, nil
	}

	// If failed, use the protocol to try to retrieve app specification.
	registrySpec, err := gh.FindSpec()
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
	registrySpecDir := appendToAbsPath(m.registriesPath, gh.VersionsDir())
	err = m.appFS.MkdirAll(string(registrySpecDir), defaultFolderPermissions)
	if err != nil {
		return nil, err
	}

	err = afero.WriteFile(m.appFS, registrySpecFile, registrySpecBytes, defaultFilePermissions)
	if err != nil {
		return nil, err
	}

	return registrySpec, nil
}
