// Copyright 2017 The kubecfg authors
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
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/blang/semver"
	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

const (
	// DefaultAPIVersion is the default ks API version to use if not specified.
	DefaultAPIVersion = "0.3.0"
	// Kind is the schema resource type.
	Kind = "ksonnet.io/app"
	// DefaultVersion is the default version of the app schema.
	DefaultVersion = "0.0.1"
)

var (
	// ErrRegistryNameInvalid is the error where a registry name is invalid.
	ErrRegistryNameInvalid = fmt.Errorf("Registry name is invalid")
	// ErrRegistryExists is the error when trying to create a registry that already exists.
	ErrRegistryExists = fmt.Errorf("Registry with name already exists")
	// ErrEnvironmentNameInvalid is the error where an environment name is invalid.
	ErrEnvironmentNameInvalid = fmt.Errorf("Environment name is invalid")
	// ErrEnvironmentExists is the error when trying to create an environment that already exists.
	ErrEnvironmentExists = fmt.Errorf("Environment with name already exists")
	// ErrEnvironmentNotExists is the error when trying to update an environment that doesn't exist.
	ErrEnvironmentNotExists = fmt.Errorf("Environment with name doesn't exist")
)

var (
	compatibleAPIRangeStrings = []string{
		">= 0.1.0 <= 0.3.0",
	}
	compatibleAPIRanges = mustCompileRanges()
)

func mustCompileRanges() []semver.Range {
	result := make([]semver.Range, 0, len(compatibleAPIRangeStrings))
	for _, s := range compatibleAPIRangeStrings {
		result = append(result, semver.MustParseRange(s))
	}
	return result
}

// Spec defines all the ksonnet project metadata. This includes details such as
// the project name, authors, environments, and registries.
type Spec = Spec030

// RepositorySpec defines the spec for the upstream repository of this project.
type RepositorySpec = RepositorySpec030

// RegistryConfig defines the spec for a registry. A registry is a collection
// of library parts.
type RegistryConfig = RegistryConfig030

// RegistryConfigs is a map of the registry name to a RegistryConfig.
type RegistryConfigs = RegistryConfigs030

// EnvironmentConfigs contains one or more EnvironmentConfig.
type EnvironmentConfigs = EnvironmentConfigs030

// EnvironmentConfig contains the specification for ksonnet environments.
type EnvironmentConfig = EnvironmentConfig030

// EnvironmentDestinationSpec contains the specification for the cluster
// address that the environment points to.
type EnvironmentDestinationSpec = EnvironmentDestinationSpec030

// LibraryConfig is the specification for a library part.
type LibraryConfig = LibraryConfig030

// LibraryConfigs is a mapping of a library configurations by name.
type LibraryConfigs = LibraryConfigs030

// ContributorSpec is a specification for the project contributors.
type ContributorSpec = ContributorSpec030

// ContributorSpecs is a list of 0 or more contributors.
type ContributorSpecs = ContributorSpecs030

type upgrader interface {
	Upgrade(fs afero.Fs, from semver.Version, fromSchema interface{}) (interface{}, error)
}

// Read will return the specification for a ksonnet application.
func read(fs afero.Fs, root string) (*Spec, error) {
	log.Debugf("loading application configuration from %s", root)

	appConfig, err := afero.ReadFile(fs, specPath(root))
	if err != nil {
		return nil, err
	}

	var base specBase

	err = yaml.Unmarshal(appConfig, &base)
	if err != nil {
		return nil, err
	}

	// Run migrations up to current version
	spec, err := applyMigrations(fs, root, base)
	if err != nil {
		return nil, errors.Wrap(err, "applying migrations")
	}

	return spec, nil
}

func applyMigrations(fs afero.Fs, root string, base specBase) (*Spec, error) {
	m := NewMigrator(fs, root)
	untyped, err := m.Load(base.APIVersion, true)
	if err != nil {
		return nil, err
	}

	spec, ok := untyped.(*Spec)
	if !ok {
		return nil, errors.Errorf("unexpected type after migration: %T", untyped)
	}

	return spec, nil
}

// Write writes the provided spec to file system.
func write(fs afero.Fs, appRoot string, spec *Spec) error {
	appConfig, err := yaml.Marshal(&spec)
	if err != nil {
		return errors.Wrap(err, "convert app configuration to YAML")
	}

	log.Debugf("writing %s", specPath(appRoot))
	if err = afero.WriteFile(fs, specPath(appRoot), appConfig, DefaultFilePermissions); err != nil {
		return errors.Wrap(err, "write app.yaml")
	}

	return nil
}

func specPath(appRoot string) string {
	return filepath.Join(appRoot, appYamlName)
}

func libName(name string) string {
	segments := strings.Split(name, "/")
	if len(segments) < 2 {
		return name
	}
	return segments[len(segments)-1]
}

func qualifyLibName(registry, name string) string {
	if registry == "" {
		return name
	}

	return fmt.Sprintf("%s/%s", registry, name)
}

var (
	errInvalidSpec = fmt.Errorf("package name should be in the form `<registry>/<library>@<version>`")
	reDescriptor   = regexp.MustCompile(`^([A-Za-z0-9\-]+)(\/[^@]+)?(@[^@]+)?$`)
)

// parseLibraryConfig parses a library identifier into its components
// <registry>/<name>@<version>
// See pkg.Parse().
func parseLibraryConfig(id string) (LibraryConfig, error) {
	var registry, name, version string

	matches := reDescriptor.FindStringSubmatch(id)
	if len(matches) == 0 {
		return LibraryConfig{}, errInvalidSpec
	}

	if matches[2] == "" {
		// No registry
		name = strings.TrimPrefix(matches[1], "/")
	} else {
		// Registry and name
		registry = matches[1]
		name = strings.TrimPrefix(matches[2], "/")
	}

	version = strings.TrimPrefix(matches[3], "@")

	return LibraryConfig{
		Registry: registry,
		Name:     name,
		Version:  version,
	}, nil
}
