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

	"github.com/ghodss/yaml"
)

const (
	DefaultAPIVersion = "0.0.1"
	Kind              = "ksonnet.io/app"
	DefaultVersion    = "0.0.1"
)

var ErrRegistryNameInvalid = fmt.Errorf("Registry name is invalid")
var ErrRegistryExists = fmt.Errorf("Registry with name already exists")

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
	Libraries    LibraryRefSpecs  `json:"libraries,omitempty"`
	License      string           `json:"license,omitempty"`
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
