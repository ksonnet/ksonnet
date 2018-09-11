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

	"github.com/blang/semver"
	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// Spec010 defines all the ksonnet project metadata. This includes details such as
// the project name, authors, environments, and registries.
type Spec010 struct {
	APIVersion   string                `json:"apiVersion,omitempty"`
	Kind         string                `json:"kind,omitempty"`
	Name         string                `json:"name,omitempty"`
	Version      string                `json:"version,omitempty"`
	Description  string                `json:"description,omitempty"`
	Authors      []string              `json:"authors,omitempty"`
	Contributors ContributorSpecs010   `json:"contributors,omitempty"`
	Repository   *RepositorySpec010    `json:"repository,omitempty"`
	Bugs         string                `json:"bugs,omitempty"`
	Keywords     []string              `json:"keywords,omitempty"`
	Registries   RegistryConfigs010    `json:"registries,omitempty"`
	Environments EnvironmentConfigs010 `json:"environments,omitempty"`
	Libraries    LibraryConfigs010     `json:"libraries,omitempty"`
	License      string                `json:"license,omitempty"`
}

// RepositorySpec010 defines the spec for the upstream repository of this project.
type RepositorySpec010 struct {
	Type string `json:"type"`
	URI  string `json:"uri"`
}

// RegistryConfig010 defines the spec for a registry. A registry is a collection
// of library parts.
type RegistryConfig010 struct {
	// Name is the user defined name of a registry.
	Name string `json:"-"`
	// Protocol is the registry protocol for this registry. Currently supported
	// values are `github`, `fs`, `helm`.
	Protocol string `json:"protocol"`
	// URI is the location of the registry.
	URI string `json:"uri"`

	isOverride bool
}

// RegistryConfigs010 is a map of the registry name to a RegistryConfig.
type RegistryConfigs010 map[string]*RegistryConfig010

// UnmarshalJSON implements the json.Unmarshaler interface.
// Our goal is to populate the Name field of RegistryConfig
// objects according to they key name in the registries map.
func (r *RegistryConfigs010) UnmarshalJSON(b []byte) error {
	registries := make(map[string]*RegistryConfig010)
	if err := json.Unmarshal(b, &registries); err != nil {
		return err
	}

	// Set Name fields according to map keys
	result := RegistryConfigs010{}
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

// EnvironmentConfigs010 contains one or more EnvironmentConfig.
type EnvironmentConfigs010 map[string]*EnvironmentConfig010

// UnmarshalJSON implements the json.Unmarshaler interface.
// Our goal is to populate the Name field of EnvironmentConfig
// objects according to they key name in the environments map.
func (e *EnvironmentConfigs010) UnmarshalJSON(b []byte) error {
	envs := make(map[string]*EnvironmentConfig010)
	if err := json.Unmarshal(b, &envs); err != nil {
		return err
	}

	// Set Name fields according to map keys
	result := EnvironmentConfigs010{}
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

// EnvironmentConfig010 contains the specification for ksonnet environments.
type EnvironmentConfig010 struct {
	// Name is the user defined name of an environment
	Name string `json:"-"`
	// KubernetesVersion is the kubernetes version the targeted cluster is
	// running on.
	KubernetesVersion string `json:"k8sVersion"`
	// Path is the relative project path containing metadata for this
	// environment.
	Path string `json:"path"`
	// Destination stores the cluster address that this environment points to.
	Destination *EnvironmentDestinationSpec010 `json:"destination"`
	// Targets contain the relative component paths that this environment
	// wishes to deploy on it's destination.
	Targets []string `json:"targets,omitempty"`

	isOverride bool
}

// EnvironmentDestinationSpec010 contains the specification for the cluster
// address that the environment points to.
type EnvironmentDestinationSpec010 struct {
	// Server is the Kubernetes server that the cluster is running on.
	Server string `json:"server"`
	// Namespace is the namespace of the Kubernetes server that targets should
	// be deployed to. This is "default", if not specified.
	Namespace string `json:"namespace"`
}

// LibraryConfig010 is the specification for a library part.
type LibraryConfig010 struct {
	Name       string             `json:"name"`
	Registry   string             `json:"registry"`
	GitVersion *GitVersionSpec010 `json:"gitVersion,omitempty"`
}

// GitVersionSpec010 is the specification for a Registry's Git Version.
type GitVersionSpec010 struct {
	RefSpec   string `json:"refSpec"`
	CommitSHA string `json:"commitSha"`
}

// LibraryConfigs010 is a mapping of a library configurations by name.
type LibraryConfigs010 map[string]*LibraryConfig010

// libraryConfigs is an alias that allows us to leverage default JSON encoding
// in our custom MarshalJSON handler without triggering infinite recursion.
type libraryConfigs010 LibraryConfigs010

// UnmarshalJSON implements the json.Unmarshaler interface.
// We implement some compatibility conversions.
func (l *LibraryConfigs010) UnmarshalJSON(b []byte) error {
	var cfgs map[string]*LibraryConfig010

	if err := json.Unmarshal(b, &cfgs); err != nil {
		return err
	}

	result := LibraryConfigs010{}
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

// MarshalJSON implements the json.Unmarshaler interface.
// We implement some compatibility conversions.
func (l *LibraryConfigs010) MarshalJSON() ([]byte, error) {
	lc := make(map[string]*LibraryConfig010)

	for k, v := range *l {
		if v == nil {
			continue
		}

		v.Name = k
		lc[k] = v
	}

	return json.Marshal(lc)
}

// libraryConfig is an alias that allows us to leverage default JSON decoding
// in our custom UnmarshalJSON handler without triggering infinite recursion.
type libraryConfig010 LibraryConfig010

// ContributorSpec010 is a specification for the project contributors.
type ContributorSpec010 struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// ContributorSpecs010 is a list of 0 or more contributors.
type ContributorSpecs010 []*ContributorSpec010

// Marshal converts a app.Spec into bytes for file writing.
func (s *Spec010) Marshal() ([]byte, error) {
	return yaml.Marshal(s)
}

// RegistryConfig returns a populated RegistryConfig given a registry name.
func (s *Spec010) RegistryConfig(name string) (*RegistryConfig010, bool) {
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
func (s *Spec010) AddRegistryConfig(cfg *RegistryConfig010) error {
	if cfg.Name == "" {
		return ErrRegistryNameInvalid
	}

	if _, exists := s.Registries[cfg.Name]; exists {
		return ErrRegistryExists
	}

	s.Registries[cfg.Name] = cfg
	return nil
}

type spec010 Spec010

// UnmarshalJSON implements the json.Unmarshaler interface.
func (s *Spec010) UnmarshalJSON(b []byte) error {
	var r spec010
	if err := json.Unmarshal(b, &r); err != nil {
		return err
	}

	if r.Contributors == nil {
		r.Contributors = ContributorSpecs010{}
	}

	if r.Registries == nil {
		r.Registries = RegistryConfigs010{}
	}

	if r.Libraries == nil {
		r.Libraries = LibraryConfigs010{}
	}

	if r.Environments == nil {
		r.Environments = EnvironmentConfigs010{}
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

	*s = Spec010(r)
	return nil
}
