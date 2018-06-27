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
	"github.com/ksonnet/ksonnet/pkg/helm"
	"github.com/pkg/errors"
)

// Add adds a registry with `name`, `protocol`, and `uri` to
// the current ksonnet application.
func Add(a app.App, protocol Protocol, name string, uri string, isOverride bool) (*Spec, error) {
	var r Registry
	var err error

	initSpec := &app.RegistryConfig{
		Name:     name,
		Protocol: string(protocol),
		URI:      uri,
	}

	switch protocol {
	case ProtocolGitHub:
		r, err = githubFactory(a, initSpec)
	case ProtocolFilesystem:
		r, err = NewFs(a, initSpec)
	case ProtocolHelm:
		var hc *helm.HTTPClient
		hc, err = helm.NewHTTPClient(initSpec.URI, nil)
		if err != nil {
			return nil, errors.Wrap(err, "initializing helm HTTP client")
		}
		r, err = helmFactory(a, initSpec, hc)
	default:
		return nil, errors.Errorf("invalid registry protocol %q", protocol)
	}

	if err != nil {
		return nil, errors.Wrap(err, "adding registry")
	}

	if ok, err := r.ValidateURI(uri); err != nil || !ok {
		return nil, errors.Wrap(err, "validating registry URL")
	}

	err = a.AddRegistry(r.MakeRegistryConfig(), isOverride)
	if err != nil {
		return nil, err
	}

	// Retrieve the contents of registry.
	registrySpec, err := r.FetchRegistrySpec()
	if err != nil {
		return nil, errors.Wrap(err, "caching registry")
	}

	return registrySpec, nil
}
