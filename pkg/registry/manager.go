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
	"net/http"
	"path/filepath"

	"github.com/ksonnet/ksonnet/pkg/app"
	"github.com/ksonnet/ksonnet/pkg/helm"
	"github.com/ksonnet/ksonnet/pkg/util/github"
	"github.com/pkg/errors"
)

// Locate locates a registry given a spec.
func Locate(a app.App, spec *app.RegistryConfig, httpClient *http.Client) (Registry, error) {
	switch Protocol(spec.Protocol) {
	case ProtocolGitHub:
		var ghc = github.NewGitHub(httpClient)
		return githubFactory(a, spec, GitHubClient(ghc))
	case ProtocolFilesystem:
		return NewFs(a, spec)
	case ProtocolHelm:
		client, err := helm.NewHTTPClient(spec.URI, httpClient)
		if err != nil {
			return nil, err
		}
		return NewHelm(a, spec, client, nil)
	default:
		return nil, errors.Errorf("invalid registry protocol %q", spec.Protocol)
	}
}

// registryCacheRoot returns the root path for registry caches
// TODO: add this to App
func registryCacheRoot(a app.App) string {
	return filepath.Join(a.Root(), ".ksonnet", "registries")
}

// registrySpecFilePath returns the path for provided registry object's cached spec file
func registrySpecFilePath(a app.App, r Registry) string {
	path := r.RegistrySpecFilePath()
	if filepath.IsAbs(path) {
		return path
	}

	return filepath.Join(registryCacheRoot(a), path)
}

// List returns a list of alphabetically sorted Registries.
func List(ksApp app.App, httpClient *http.Client) ([]Registry, error) {
	var registries []Registry
	appRegistries, err := ksApp.Registries()
	if err != nil {
		return nil, err
	}
	for name, regRef := range appRegistries {
		regRef.Name = name
		r, err := Locate(ksApp, regRef, httpClient)
		if err != nil {
			return nil, err
		}
		registries = append(registries, r)
	}

	return registries, nil
}
