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

// Spec020 defines all the ksonnet project metadata. This includes details such as
// the project name, authors, environments, and registries.
type Spec020 struct {
	APIVersion   string                `json:"apiVersion,omitempty"`
	Kind         string                `json:"kind,omitempty"`
	Name         string                `json:"name,omitempty"`
	Version      string                `json:"version,omitempty"`
	Description  string                `json:"description,omitempty"`
	Authors      []string              `json:"authors,omitempty"`
	Contributors ContributorSpecs020   `json:"contributors,omitempty"`
	Repository   *RepositorySpec020    `json:"repository,omitempty"`
	Bugs         string                `json:"bugs,omitempty"`
	Keywords     []string              `json:"keywords,omitempty"`
	Registries   RegistryConfigs020    `json:"registries,omitempty"`
	Environments EnvironmentConfigs020 `json:"environments,omitempty"`
	Libraries    LibraryConfigs020     `json:"libraries,omitempty"`
	License      string                `json:"license,omitempty"`
}

// RepositorySpec020 defines the spec for the upstream repository of this project.
type RepositorySpec020 struct {
	Type string `json:"type"`
	URI  string `json:"uri"`
}

// RegistryConfig020 defines the spec for a registry. A registry is a collection
// of library parts.
type RegistryConfig020 struct {
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
func (r *RegistryConfig020) IsOverride() bool {
	return r.isOverride
}

// RegistryConfigs020 is a map of the registry name to a RegistryConfig.
type RegistryConfigs020 map[string]*RegistryConfig020

// UnmarshalJSON implements the json.Unmarshaler interface.
// Our goal is to populate the Name field of RegistryConfig
// objects according to they key name in the registries map.
func (r *RegistryConfigs020) UnmarshalJSON(b []byte) error {
	registries := make(map[string]*RegistryConfig020)
	if err := json.Unmarshal(b, &registries); err != nil {
		return err
	}

	// Set Name fields according to map keys
	result := RegistryConfigs020{}
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

// EnvironmentConfigs020 contains one or more EnvironmentConfig.
type EnvironmentConfigs020 map[string]*EnvironmentConfig020

// UnmarshalJSON implements the json.Unmarshaler interface.
// Our goal is to populate the Name field of EnvironmentConfig
// objects according to they key name in the environments map.
func (e *EnvironmentConfigs020) UnmarshalJSON(b []byte) error {
	envs := make(map[string]*EnvironmentConfig020)
	if err := json.Unmarshal(b, &envs); err != nil {
		return err
	}

	// Set Name fields according to map keys
	result := EnvironmentConfigs020{}
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

// EnvironmentConfig020 contains the specification for ksonnet environments.
type EnvironmentConfig020 struct {
	// Name is the user defined name of an environment
	Name string `json:"-"`
	// KubernetesVersion is the kubernetes version the targeted cluster is
	// running on.
	KubernetesVersion string `json:"k8sVersion"`
	// Path is the relative project path containing metadata for this
	// environment.
	Path string `json:"path"`
	// Destination stores the cluster address that this environment points to.
	Destination *EnvironmentDestinationSpec020 `json:"destination"`
	// Targets contain the relative component paths that this environment
	// wishes to deploy on it's destination.
	Targets []string `json:"targets,omitempty"`
	// Libraries specifies versioned libraries specifically used by this environment.
	Libraries LibraryConfigs020 `json:"libraries,omitempty"`

	isOverride bool
}

// MakePath return the absolute path to the environment directory.
func (e *EnvironmentConfig020) MakePath(rootPath string) string {
	return filepath.Join(
		rootPath,
		EnvironmentDirName,
		filepath.FromSlash(e.Path))
}

// IsOverride is true if this EnvironmentConfig is an override.
func (e *EnvironmentConfig020) IsOverride() bool {
	return e.isOverride
}

// EnvironmentDestinationSpec020 contains the specification for the cluster
// address that the environment points to.
type EnvironmentDestinationSpec020 struct {
	// Server is the Kubernetes server that the cluster is running on.
	Server string `json:"server"`
	// Namespace is the namespace of the Kubernetes server that targets should
	// be deployed to. This is "default", if not specified.
	Namespace string `json:"namespace"`
}

// LibraryConfig020 is the specification for a library part.
type LibraryConfig020 struct {
	Name     string `json:"name"`
	Registry string `json:"registry"`
	Version  string `json:"version"`
}

// GitVersionSpec020 is the specification for a Registry's Git Version.
type GitVersionSpec020 struct {
	RefSpec   string `json:"refSpec"`
	CommitSHA string `json:"commitSha"`
}

// LibraryConfigs020 is a mapping of a library configurations by name.
type LibraryConfigs020 map[string]*LibraryConfig020

// libraryConfigs is an alias that allows us to leverage default JSON encoding
// in our custom MarshalJSON handler without triggering infinite recursion.
type libraryConfigs020 LibraryConfigs020

// UnmarshalJSON implements the json.Unmarshaler interface.
// We implement some compatibility conversions.
func (l *LibraryConfigs020) UnmarshalJSON(b []byte) error {
	var cfgs map[string]*LibraryConfig020

	if err := json.Unmarshal(b, &cfgs); err != nil {
		return err
	}

	result := LibraryConfigs020{}
	for k, v := range cfgs {
		if v == nil {
			continue
		}

		v.Name = k
		result[k] = v
	}

	*l = result
	return nil
}

// libraryConfig is an alias that allows us to leverage default JSON decoding
// in our custom UnmarshalJSON handler without triggering infinite recursion.
type libraryConfig020 LibraryConfig020

func (l LibraryConfig020) String() string {
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

// ContributorSpec020 is a specification for the project contributors.
type ContributorSpec020 struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// ContributorSpecs020 is a list of 0 or more contributors.
type ContributorSpecs020 []*ContributorSpec020

// Marshal converts a app.Spec into bytes for file writing.
func (s *Spec020) Marshal() ([]byte, error) {
	return yaml.Marshal(s)
}

// RegistryConfig returns a populated RegistryConfig given a registry name.
func (s *Spec020) RegistryConfig(name string) (*RegistryConfig020, bool) {
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
func (s *Spec020) AddRegistryConfig(cfg *RegistryConfig020) error {
	if cfg.Name == "" {
		return ErrRegistryNameInvalid
	}

	if _, exists := s.Registries[cfg.Name]; exists {
		return ErrRegistryExists
	}

	s.Registries[cfg.Name] = cfg
	return nil
}

type spec020 Spec020

// UnmarshalJSON implements the json.Unmarshaler interface.
func (s *Spec020) UnmarshalJSON(b []byte) error {
	var r spec020
	if err := json.Unmarshal(b, &r); err != nil {
		return err
	}

	if r.Contributors == nil {
		r.Contributors = ContributorSpecs020{}
	}

	if r.Registries == nil {
		r.Registries = RegistryConfigs020{}
	}

	if r.Libraries == nil {
		r.Libraries = LibraryConfigs020{}
	}

	if r.Environments == nil {
		r.Environments = EnvironmentConfigs020{}
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

	*s = Spec020(r)
	return nil
}
