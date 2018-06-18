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

package registry

import (
	"github.com/ksonnet/ksonnet/pkg/app"
	"github.com/ksonnet/ksonnet/pkg/parts"
)

// Protocol is the protocol for a registry.
type Protocol string

func (p Protocol) String() string {
	return string(p)
}

const (
	// ProtocolFilesystem is the protocol for file system based registries.
	ProtocolFilesystem Protocol = "fs"
	// ProtocolGitHub is the protocol for GitHub based registries.
	ProtocolGitHub Protocol = "github"
	// ProtocolHelm is the protocol for Helm based registries.
	ProtocolHelm Protocol = "helm"

	registryYAMLFile = "registry.yaml"
	partsYAMLFile    = "parts.yaml"
)

// ResolveFile resolves files found when searching a part.
type ResolveFile func(relPath string, contents []byte) error

// ResolveDirectory resolves directories when searching a part.
type ResolveDirectory func(relPath string) error

// Registry is a Registry
type Registry interface {
	RegistrySpecDir() string
	RegistrySpecFilePath() string
	SpecFetcher
	MakeRegistryRefSpec() *app.RegistryRefSpec
	Resolver
	Name() string
	Protocol() Protocol
	URI() string
	IsOverride() bool
	CacheRoot(name, relPath string) (string, error)

	Updater
	Validator
}

type SpecFetcher interface {
	FetchRegistrySpec() (*Spec, error)
}

type Resolver interface {
	ResolveLibrarySpec(libID, libRefSpec string) (*parts.Spec, error)
	ResolveLibrary(libID, libAlias, version string, onFile ResolveFile, onDir ResolveDirectory) (*parts.Spec, *app.LibraryRefSpec, error)
}

// Updater is an interface for updating an existing registry
type Updater interface {
	Update(version string) (newVersion string, err error)
}

// Validator is an interface for validating a registry URI
type Validator interface {
	ValidateURI(uri string) (bool, error)
}
