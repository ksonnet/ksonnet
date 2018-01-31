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

	"github.com/blang/semver"
	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
)

const (
	DefaultAPIVersion = "0.0.1"
	Kind              = "ksonnet.io/app"
	DefaultVersion    = "0.0.1"
)

var (
	ErrRegistryNameInvalid = fmt.Errorf("Registry name is invalid")
	ErrRegistryExists      = fmt.Errorf("Registry with name already exists")
	// ErrEnvironmentNameInvalid is the error where an environment name is invalid.
	ErrEnvironmentNameInvalid = fmt.Errorf("Environment name is invalid")
	// ErrEnvironmentExists is the error when trying to create an environment that already exists.
	ErrEnvironmentExists = fmt.Errorf("Environment with name already exists")
)

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
	Registries   RegistryRefSpecs `json:"registries,omitempty"`
	Environments EnvironmentSpecs `json:"environments,omitempty"`
	Libraries    LibraryRefSpecs  `json:"libraries,omitempty"`
	License      string           `json:"license,omitempty"`
}

type RepositorySpec struct {
	Type string `json:"type"`
	URI  string `json:"uri"`
}

type RegistryRefSpec struct {
	Name       string          `json:"-"`
	Protocol   string          `json:"protocol"`
	URI        string          `json:"uri"`
	GitVersion *GitVersionSpec `json:"gitVersion"`
}

type RegistryRefSpecs map[string]*RegistryRefSpec

// EnvironmentSpecs contains one or more EnvironmentSpec.
type EnvironmentSpecs map[string]*EnvironmentSpec

// EnvironmentSpec contains the specification for ksonnet environments.
//
// KubernetesVersion: The Kubernetes version the target cluster is running on.
// Destinations:      One or more cluster addresses that this environment
//                    points to.
// Targets:           The relative component paths that this environment wishes
//                    to deploy onto it's destinations.
type EnvironmentSpec struct {
	Name              string                      `json:"-"`
	KubernetesVersion string                      `json:"k8sVersion"`
	Destinations      EnvironmentDestinationSpecs `json:"destinations"`
	Targets           []string                    `json:"targets"`
}

// EnvironmentDestinationSpecs contains one or more EnvironmentDestinationSpec.
type EnvironmentDestinationSpecs []*EnvironmentDestinationSpec

// EnvironmentDestinationSpec contains the specification for the cluster
// addresses that the environment points to.
//
// Server:    The Kubernetes server that the cluster is running on.
// Namespace: The namespace of the Kubernetes server that targets should
//            be deployed to. This is "default", by default.
type EnvironmentDestinationSpec struct {
	Server    string `json:"server"`
	Namespace string `json:"namespace"`
}

type LibraryRefSpec struct {
	Name       string          `json:"name"`
	Registry   string          `json:"registry"`
	GitVersion *GitVersionSpec `json:"gitVersion"`
}

type GitVersionSpec struct {
	RefSpec   string `json:"refSpec"`
	CommitSHA string `json:"commitSha"`
}

type LibraryRefSpecs map[string]*LibraryRefSpec

type ContributorSpec struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type ContributorSpecs []*ContributorSpec

func Unmarshal(bytes []byte) (*Spec, error) {
	schema := Spec{}
	err := yaml.Unmarshal(bytes, &schema)
	if err != nil {
		return nil, err
	}

	if err = schema.validate(); err != nil {
		return nil, err
	}

	return &schema, nil
}

func (s *Spec) Marshal() ([]byte, error) {
	return yaml.Marshal(s)
}

func (s *Spec) GetRegistryRef(name string) (*RegistryRefSpec, bool) {
	registryRefSpec, ok := s.Registries[name]
	if ok {
		// Populate name, which we do not include in the deserialization
		// process.
		registryRefSpec.Name = name
	}
	return registryRefSpec, ok
}

func (s *Spec) AddRegistryRef(registryRefSpec *RegistryRefSpec) error {
	if registryRefSpec.Name == "" {
		return ErrRegistryNameInvalid
	}

	_, registryRefExists := s.Registries[registryRefSpec.Name]
	if registryRefExists {
		return ErrRegistryExists
	}

	s.Registries[registryRefSpec.Name] = registryRefSpec
	return nil
}

func (s *Spec) validate() error {
	compatVer, _ := semver.Make(DefaultAPIVersion)
	ver, err := semver.Make(s.APIVersion)
	if err != nil {
		return errors.Wrap(err, "Failed to parse version in app spec")
	} else if compatVer.Compare(ver) != 0 {
		return fmt.Errorf(
			"Current app uses unsupported spec version '%s' (this client only supports %s)",
			s.APIVersion,
			DefaultAPIVersion)
	}

	return nil
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
