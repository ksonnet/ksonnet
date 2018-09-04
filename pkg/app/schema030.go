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
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/blang/semver"
	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// Spec030 defines all the ksonnet project metadata. This includes details such as
// the project name, authors, environments, and registries.
type Spec030 struct {
	APIVersion   string                `json:"apiVersion,omitempty"`
	Kind         string                `json:"kind,omitempty"`
	Name         string                `json:"name,omitempty"`
	Version      string                `json:"version,omitempty"`
	Description  string                `json:"description,omitempty"`
	Authors      []string              `json:"authors,omitempty"`
	Contributors ContributorSpecs030   `json:"contributors,omitempty"`
	Repository   *RepositorySpec030    `json:"repository,omitempty"`
	Bugs         string                `json:"bugs,omitempty"`
	Keywords     []string              `json:"keywords,omitempty"`
	Registries   RegistryConfigs030    `json:"registries,omitempty"`
	Environments EnvironmentConfigs030 `json:"environments,omitempty"`
	Libraries    LibraryConfigs030     `json:"libraries,omitempty"`
	License      string                `json:"license,omitempty"`
}

// RepositorySpec030 defines the spec for the upstream repository of this project.
type RepositorySpec030 struct {
	Type string `json:"type"`
	URI  string `json:"uri"`
}

// RegistryConfig030 defines the spec for a registry. A registry is a collection
// of library parts.
type RegistryConfig030 struct {
	// Name is the user defined name of a registry.
	Name string `json:"-"`
	// Protocol is the registry protocol for this registry. Currently supported
	// values are `github`, `fs`, `helm`.
	Protocol string `json:"protocol"`
	// URI is the location of the registry.
	URI string `json:"uri"`

	isOverride bool
}

// IsOverride is true if this RegistryConfig is an override.
func (r *RegistryConfig030) IsOverride() bool {
	return r.isOverride
}

// RegistryConfigs030 is a map of the registry name to a RegistryConfig.
type RegistryConfigs030 map[string]*RegistryConfig030

// UnmarshalJSON implements the json.Unmarshaler interface.
// Our goal is to populate the Name field of RegistryConfig
// objects according to they key name in the registries map.
func (r *RegistryConfigs030) UnmarshalJSON(b []byte) error {
	registries := make(map[string]*RegistryConfig030)
	if err := json.Unmarshal(b, &registries); err != nil {
		return err
	}

	// Set Name fields according to map keys
	result := RegistryConfigs030{}
	for k, v := range registries {
		if v == nil {
			continue
		}
		v.Name = k
		result[k] = v
	}

	*r = result
	return nil
}

// EnvironmentConfigs030 contains one or more EnvironmentConfig.
type EnvironmentConfigs030 map[string]*EnvironmentConfig030

// UnmarshalJSON implements the json.Unmarshaler interface.
// Our goal is to populate the Name field of EnvironmentConfig
// objects according to they key name in the environments map.
func (e *EnvironmentConfigs030) UnmarshalJSON(b []byte) error {
	envs := make(map[string]*EnvironmentConfig030)
	if err := json.Unmarshal(b, &envs); err != nil {
		return err
	}

	// Set Name fields according to map keys
	result := EnvironmentConfigs030{}
	for k, v := range envs {
		if v == nil {
			continue
		}
		v.Name = k
		result[k] = v
	}

	*e = result
	return nil
}

// EnvironmentConfig030 contains the specification for ksonnet environments.
type EnvironmentConfig030 struct {
	// Name is the user defined name of an environment
	Name string `json:"-"`
	// KubernetesVersion is the kubernetes version the targeted cluster is
	// running on.
	KubernetesVersion string `json:"k8sVersion"`
	// Path is the relative project path containing metadata for this
	// environment.
	Path string `json:"path"`
	// Destination stores the cluster address that this environment points to.
	Destination *EnvironmentDestinationSpec030 `json:"destination"`
	// Targets contain the relative component paths that this environment
	// wishes to deploy on it's destination.
	Targets []string `json:"targets,omitempty"`
	// Libraries specifies versioned libraries specifically used by this environment.
	Libraries LibraryConfigs030 `json:"libraries,omitempty"`

	isOverride bool
}

// MakePath return the absolute path to the environment directory.
func (e *EnvironmentConfig030) MakePath(rootPath string) string {
	return filepath.Join(
		rootPath,
		EnvironmentDirName,
		filepath.FromSlash(e.Path))
}

// IsOverride is true if this EnvironmentConfig is an override.
func (e *EnvironmentConfig030) IsOverride() bool {
	return e.isOverride
}

// EnvironmentDestinationSpec030 contains the specification for the cluster
// address that the environment points to.
type EnvironmentDestinationSpec030 struct {
	// Server is the Kubernetes server that the cluster is running on.
	Server string `json:"server"`
	// Namespace is the namespace of the Kubernetes server that targets should
	// be deployed to. This is "default", if not specified.
	Namespace string `json:"namespace"`
}

// LibraryConfig030 is the specification for a library part.
type LibraryConfig030 struct {
	Name     string `json:"name"`
	Registry string `json:"registry"`
	Version  string `json:"version"`
}

// GitVersionSpec030 is the specification for a Registry's Git Version.
type GitVersionSpec030 struct {
	RefSpec   string `json:"refSpec"`
	CommitSHA string `json:"commitSha"`
}

// LibraryConfigs030 is a mapping of a library configurations by name.
type LibraryConfigs030 map[string]*LibraryConfig030

// libraryConfigs is an alias that allows us to leverage default JSON encoding
// in our custom MarshalJSON handler without triggering infinite recursion.
type libraryConfigs030 LibraryConfigs030

// UnmarshalJSON implements the json.Unmarshaler interface.
// We implement some compatibility conversions.
func (l *LibraryConfigs030) UnmarshalJSON(b []byte) error {
	var cfgs libraryConfigs030

	if err := json.Unmarshal(b, &cfgs); err != nil {
		return err
	}

	result := LibraryConfigs030{}
	for k, v := range cfgs {
		if v == nil {
			continue
		}

		name := libName(k)
		qualifiedName := qualifyLibName(v.Registry, name)

		v.Name = name
		result[qualifiedName] = v
	}

	*l = result
	return nil
}

// MarshalJSON implements the json.Marshaler interface.
// We implement some compatibility conversions.
func (l *LibraryConfigs030) MarshalJSON() ([]byte, error) {
	lc := make(map[string]*LibraryConfig030)

	for k, v := range *l {
		if v == nil {
			continue
		}

		name := libName(k)
		qualifiedName := qualifyLibName(v.Registry, name)

		v.Name = name
		lc[qualifiedName] = v
	}

	return json.Marshal(lc)
}

func (l LibraryConfig030) String() string {
	switch {
	case l.Registry != "" && l.Version != "":
		return fmt.Sprintf("%s/%s@%s", l.Registry, l.Name, l.Version)
	case l.Registry != "" && l.Version == "":
		return fmt.Sprintf("%s/%s", l.Registry, l.Name)
	case l.Registry == "" && l.Version != "":
		return fmt.Sprintf("%s@%s", l.Name, l.Version)
	case l.Registry == "" && l.Version == "":
		return l.Name
	default:
		// Not sure which case we missed, just default to verbose
		return fmt.Sprintf("%s/%s@%s", l.Registry, l.Name, l.Version)
	}
}

// ContributorSpec030 is a specification for the project contributors.
type ContributorSpec030 struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// ContributorSpecs030 is a list of 0 or more contributors.
type ContributorSpecs030 []*ContributorSpec030

// Marshal converts a app.Spec into bytes for file writing.
func (s *Spec030) Marshal() ([]byte, error) {
	return yaml.Marshal(s)
}

// RegistryConfig returns a populated RegistryConfig given a registry name.
func (s *Spec030) RegistryConfig(name string) (*RegistryConfig030, bool) {
	cfg, ok := s.Registries[name]
	if ok {
		// Verify map name matches the name in configuration. These should always match.
		if cfg.Name != name {
			log.WithField("action", "app.Spec.RegistryConfig").Warnf("registry configuration name mismatch: %v vs. %v", cfg.Name, name)
			cfg.Name = name
		}
	}
	return cfg, ok
}

// AddRegistryConfig adds the RegistryConfig to the app spec.
func (s *Spec030) AddRegistryConfig(cfg *RegistryConfig030) error {
	if cfg.Name == "" {
		return ErrRegistryNameInvalid
	}

	if _, exists := s.Registries[cfg.Name]; exists {
		return ErrRegistryExists
	}

	s.Registries[cfg.Name] = cfg
	return nil
}

type spec030 Spec030

// UnmarshalJSON implements the json.Unmarshaler interface.
func (s *Spec030) UnmarshalJSON(b []byte) error {
	var r spec030
	if err := json.Unmarshal(b, &r); err != nil {
		return err
	}

	if r.Contributors == nil {
		r.Contributors = ContributorSpecs030{}
	}

	if r.Registries == nil {
		r.Registries = RegistryConfigs030{}
	}

	if r.Libraries == nil {
		r.Libraries = LibraryConfigs030{}
	}

	if r.Environments == nil {
		r.Environments = EnvironmentConfigs030{}
	}

	if r.APIVersion == "0.0.0" {
		return errors.New("invalid version")
	}

	ver, err := semver.Make(r.APIVersion)
	if err != nil {
		return errors.Wrap(err, "Failed to parse version in app spec")
	}

	var compatible bool
	for _, compatRange := range compatibleAPIRanges {
		if compatRange(ver) {
			compatible = true
		}
	}

	if !compatible {
		return fmt.Errorf(
			"Current app uses unsupported spec version '%s' (this client only supports %s)",
			r.APIVersion,
			compatibleAPIRangeStrings)
	}

	*s = Spec030(r)
	return nil
}

// GetEnvironmentConfigs returns all environment specifications.
// TODO: Consider returning copies instead of originals
func (s *Spec030) GetEnvironmentConfigs() EnvironmentConfigs {
	for k, v := range s.Environments {
		v.Name = k
	}

	return s.Environments
}

// GetEnvironmentConfig returns the environment specification for the environment.
// TODO: Consider returning copies instead of originals
func (s *Spec030) GetEnvironmentConfig(name string) (*EnvironmentConfig030, bool) {
	env, ok := s.Environments[name]
	if ok {
		env.Name = name
	}
	return env, ok
}

// AddEnvironmentConfig adds an EnvironmentConfig to the list of EnvironmentConfigs.
// This is equivalent to registering the environment for a ksonnet app.
func (s *Spec030) AddEnvironmentConfig(env *EnvironmentConfig030) error {
	if env.Name == "" {
		return ErrEnvironmentNameInvalid
	}

	if _, ok := s.Environments[env.Name]; ok {
		return ErrEnvironmentExists
	}

	s.Environments[env.Name] = env
	return nil
}

// DeleteEnvironmentConfig removes the environment specification from the app spec.
func (s *Spec030) DeleteEnvironmentConfig(name string) error {
	delete(s.Environments, name)
	return nil
}

// UpdateEnvironmentConfig updates the environment with the provided name to the
// specified spec.
func (s *Spec030) UpdateEnvironmentConfig(name string, env *EnvironmentConfig030) error {
	if env.Name == "" {
		return ErrEnvironmentNameInvalid
	}

	_, ok := s.Environments[name]
	if !ok {
		return errors.Errorf("Environment with name %q does not exist", name)
	}

	if name != env.Name {
		if err := s.DeleteEnvironmentConfig(name); err != nil {
			return err
		}
	}

	s.Environments[env.Name] = env
	return nil
}
