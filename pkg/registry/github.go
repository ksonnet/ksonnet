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

import "github.com/ksonnet/ksonnet/metadata/app"

// GitHub is a GitHub based registry.
type GitHub struct {
	name string
	spec *app.RegistryRefSpec
}

// NewGitHub creates an instance of GitHub.
func NewGitHub(name string, spec *app.RegistryRefSpec) *GitHub {
	return &GitHub{
		name: name,
		spec: spec,
	}
}

var _ Registry = (*GitHub)(nil)

// Name is the registry name.
func (g *GitHub) Name() string {
	return g.name
}

// Protocol is the registry protocol.
func (g *GitHub) Protocol() string {
	return g.spec.Protocol
}

// URI is the registry URI.
func (g *GitHub) URI() string {
	return g.spec.URI
}
