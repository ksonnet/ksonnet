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
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/blang/semver"
	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

const (
	// DefaultAPIVersion is the default ks API version to use if not specified.
	DefaultAPIVersion = "0.1.0"
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

// Spec defines all the ksonnet project metadata. This includes details such as
// the project name, authors, environments, and registries.
type Spec struct {
	APIVersion   string           `json:"apiVersion,omitempty"`
	Kind         string           `json:"kind,omitempty"`
	Name         string           `json:"name,omitempty"`
	Version      string           `json:"version,omitempty"`
	Description  string           `json:"description,omitempty"`
	Authors      []string         `json:"authors,omitempty"`
	Contributors ContributorSpecs `json:"contributors,omitempty"`
	Repository   *RepositorySpec  `json:"repository,omitempty"`
	Bugs         string           `json:"bugs,omitempty"`
	Keywords     []string         `json:"keywords,omitempty"`
	Registries   RegistryConfigs  `json:"registries,omitempty"`
	Environments EnvironmentSpecs `json:"environments,omitempty"`
	Libraries    LibraryConfigs   `json:"libraries,omitempty"`
	License      string           `json:"license,omitempty"`
}

// Read will return the specification for a ksonnet application. It will navigate up directories
// to search for `app.yaml` and return error if it hits the root directory.
func read(fs afero.Fs, root string) (*Spec, error) {
	log.Debugf("loading application configuration from %s", root)

	appConfig, err := afero.ReadFile(fs, specPath(root))
	if err != nil {
		return nil, err
	}

	var spec Spec

	err = yaml.Unmarshal(appConfig, &spec)
	if err != nil {
		return nil, err
	}

	if err = spec.validate(); err != nil {
		return nil, err
	}

	exists, err := afero.Exists(fs, overridePath(root))
	if err != nil {
		return nil, err
	}

	if exists {
		var o Override

		overrideConfig, err := afero.ReadFile(fs, overridePath(root))
		if err != nil {
			return nil, err
		}

		err = yaml.Unmarshal(overrideConfig, &o)
		if err != nil {
			return nil, err
		}

		for k, v := range o.Environments {
			v.isOverride = true
			spec.Environments[k] = v
		}

		for k, v := range o.Registries {
			v.isOverride = true
			spec.Registries[k] = v
		}
	}

	if err := spec.validate(); err != nil {
		return nil, err
	}

	return &spec, nil
}

// Write writes the provided spec to file system.
func write(fs afero.Fs, appRoot string, spec *Spec) error {
	o := Override{
		Kind:         overrideKind,
		APIVersion:   overrideVersion,
		Environments: EnvironmentSpecs{},
		Registries:   RegistryConfigs{},
	}

	overrideKeys := map[string][]string{
		"environments": make([]string, 0),
		"registries":   make([]string, 0),
	}

	for k, v := range spec.Environments {
		if v.IsOverride() {
			o.Environments[k] = v
			overrideKeys["environments"] = append(overrideKeys["environments"], k)
		}
	}

	for k, v := range spec.Registries {
		if v.IsOverride() {
			o.Registries[k] = v
			overrideKeys["registries"] = append(overrideKeys["registries"], k)
		}
	}

	for _, k := range overrideKeys["environments"] {
		delete(spec.Environments, k)
	}

	for _, k := range overrideKeys["registries"] {
		delete(spec.Registries, k)
	}

	appConfig, err := yaml.Marshal(&spec)
	if err != nil {
		return errors.Wrap(err, "convert app configuration to YAML")
	}

	if err = afero.WriteFile(fs, specPath(appRoot), appConfig, DefaultFilePermissions); err != nil {
		return errors.Wrap(err, "write app.yaml")
	}

	if err = removeOverride(fs, appRoot); err != nil {
		return errors.Wrap(err, "clean overrides")
	}

	if o.IsDefined() {
		return SaveOverride(defaultYAMLEncoder, fs, appRoot, &o)
	}

	return nil
}

func removeOverride(fs afero.Fs, appRoot string) error {
	exists, err := afero.Exists(fs, overridePath(appRoot))
	if err != nil {
		return err
	}

	if exists {
		return fs.Remove(overridePath(appRoot))
	}

	return nil
}

func specPath(appRoot string) string {
	return filepath.Join(appRoot, appYamlName)
}

func overridePath(appRoot string) string {
	return filepath.Join(appRoot, overrideYamlName)
}

// RepositorySpec defines the spec for the upstream repository of this project.
type RepositorySpec struct {
	Type string `json:"type"`
	URI  string `json:"uri"`
}

// RegistryConfig defines the spec for a registry. A registry is a collection
// of library parts.
type RegistryConfig struct {
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
func (r *RegistryConfig) IsOverride() bool {
	return r.isOverride
}

// RegistryConfigs is a map of the registry name to a RegistryConfig.
type RegistryConfigs map[string]*RegistryConfig

// UnmarshalJSON implements the json.Unmarshaler interface.
// Our goal is to populate the Name field of RegistryConfig
// objects according to they key name in the registries map.
func (r *RegistryConfigs) UnmarshalJSON(b []byte) error {
	registries := make(map[string]*RegistryConfig)
	if err := json.Unmarshal(b, &registries); err != nil {
		return err
	}

	// Set Name fields according to map keys
	for k, v := range registries {
		v.Name = k
	}

	*r = RegistryConfigs(registries)
	return nil
}

// EnvironmentSpecs contains one or more EnvironmentSpec.
type EnvironmentSpecs map[string]*EnvironmentSpec

// EnvironmentSpec contains the specification for ksonnet environments.
type EnvironmentSpec struct {
	// Name is the user defined name of an environment
	Name string `json:"-"`
	// KubernetesVersion is the kubernetes version the targeted cluster is
	// running on.
	KubernetesVersion string `json:"k8sVersion"`
	// Path is the relative project path containing metadata for this
	// environment.
	Path string `json:"path"`
	// Destination stores the cluster address that this environment points to.
	Destination *EnvironmentDestinationSpec `json:"destination"`
	// Targets contain the relative component paths that this environment
	// wishes to deploy on it's destination.
	Targets []string `json:"targets,omitempty"`

	isOverride bool
}

// MakePath return the absolute path to the environment directory.
func (e *EnvironmentSpec) MakePath(rootPath string) string {
	return filepath.Join(
		rootPath,
		EnvironmentDirName,
		filepath.FromSlash(e.Path))
}

// IsOverride is true if this EnvironmentSpec is an override.
func (e *EnvironmentSpec) IsOverride() bool {
	return e.isOverride
}

// EnvironmentDestinationSpec contains the specification for the cluster
// address that the environment points to.
type EnvironmentDestinationSpec struct {
	// Server is the Kubernetes server that the cluster is running on.
	Server string `json:"server"`
	// Namespace is the namespace of the Kubernetes server that targets should
	// be deployed to. This is "default", if not specified.
	Namespace string `json:"namespace"`
}

// LibraryConfig is the specification for a library part.
type LibraryConfig struct {
	Name       string          `json:"name"`
	Registry   string          `json:"registry"`
	Version    string          `json:"version"`
	GitVersion *GitVersionSpec `json:"gitVersion,omitempty"`
}

// GitVersionSpec is the specification for a Registry's Git Version.
type GitVersionSpec struct {
	RefSpec   string `json:"refSpec"`
	CommitSHA string `json:"commitSha"`
}

// LibraryConfigs is a mapping of a library configurations by name.
type LibraryConfigs map[string]*LibraryConfig

// ContributorSpec is a specification for the project contributors.
type ContributorSpec struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// ContributorSpecs is a list of 0 or more contributors.
type ContributorSpecs []*ContributorSpec

// Marshal converts a app.Spec into bytes for file writing.
func (s *Spec) Marshal() ([]byte, error) {
	return yaml.Marshal(s)
}

// RegistryConfig returns a populated RegistryConfig given a registry name.
func (s *Spec) RegistryConfig(name string) (*RegistryConfig, bool) {
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
func (s *Spec) AddRegistryConfig(RegistryConfig *RegistryConfig) error {
	if RegistryConfig.Name == "" {
		return ErrRegistryNameInvalid
	}

	if _, exists := s.Registries[RegistryConfig.Name]; exists {
		return ErrRegistryExists
	}

	s.Registries[RegistryConfig.Name] = RegistryConfig
	return nil
}

func (s *Spec) validate() error {
	if s.Contributors == nil {
		s.Contributors = ContributorSpecs{}
	}

	if s.Registries == nil {
		s.Registries = RegistryConfigs{}
	}

	if s.Libraries == nil {
		s.Libraries = LibraryConfigs{}
	}

	if s.Environments == nil {
		s.Environments = EnvironmentSpecs{}
	}

	if s.APIVersion == "0.0.0" {
		return errors.New("invalid version")
	}

	compatVer, _ := semver.Make(DefaultAPIVersion)
	ver, err := semver.Make(s.APIVersion)
	if err != nil {
		return errors.Wrap(err, "Failed to parse version in app spec")
	}

	if compatVer.Compare(ver) < 0 {
		return fmt.Errorf(
			"Current app uses unsupported spec version '%s' (this client only supports %s)",
			s.APIVersion,
			DefaultAPIVersion)
	}

	return nil
}

// GetEnvironmentSpecs returns all environment specifications.
// We need to pre-populate th EnvironmentSpec name before returning.
func (s *Spec) GetEnvironmentSpecs() EnvironmentSpecs {
	for k, v := range s.Environments {
		v.Name = k
	}

	return s.Environments
}

// GetEnvironmentSpec returns the environment specification for the environment.
func (s *Spec) GetEnvironmentSpec(name string) (*EnvironmentSpec, bool) {
	environmentSpec, ok := s.Environments[name]
	if ok {
		environmentSpec.Name = name
	}
	return environmentSpec, ok
}

// AddEnvironmentSpec adds an EnvironmentSpec to the list of EnvironmentSpecs.
// This is equivalent to registering the environment for a ksonnet app.
func (s *Spec) AddEnvironmentSpec(spec *EnvironmentSpec) error {
	if spec.Name == "" {
		return ErrEnvironmentNameInvalid
	}

	_, environmentSpecExists := s.Environments[spec.Name]
	if environmentSpecExists {
		return ErrEnvironmentExists
	}

	s.Environments[spec.Name] = spec
	return nil
}

// DeleteEnvironmentSpec removes the environment specification from the app spec.
func (s *Spec) DeleteEnvironmentSpec(name string) error {
	delete(s.Environments, name)
	return nil
}

// UpdateEnvironmentSpec updates the environment with the provided name to the
// specified spec.
func (s *Spec) UpdateEnvironmentSpec(name string, spec *EnvironmentSpec) error {
	if spec.Name == "" {
		return ErrEnvironmentNameInvalid
	}

	_, environmentSpecExists := s.Environments[name]
	if !environmentSpecExists {
		return errors.Errorf("Environment with name %q does not exist", name)
	}

	if name != spec.Name {
		if err := s.DeleteEnvironmentSpec(name); err != nil {
			return err
		}
	}

	s.Environments[spec.Name] = spec
	return nil
}
