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
	"path/filepath"

	"github.com/ksonnet/ksonnet/metadata/app"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
)

// Add adds a registry with `name`, `protocol`, and `uri` to
// the current ksonnet application.
func Add(a app.App, name, protocol, uri, version string, isOverride bool) (*Spec, error) {
	var r Registry
	var err error

	initSpec := &app.RegistryRefSpec{
		Name:     name,
		Protocol: protocol,
		URI:      uri,
	}

	switch protocol {
	case ProtocolGitHub:
		r, err = githubFactory(initSpec)
	case ProtocolFilesystem:
		r, err = NewFs(a, initSpec)
	default:
		return nil, errors.Errorf("invalid registry protocol %q", protocol)
	}

	if err != nil {
		return nil, err
	}

	err = a.AddRegistry(r.MakeRegistryRefSpec(), isOverride)
	if err != nil {
		return nil, err
	}

	// Retrieve the contents of registry.
	registrySpec, err := getOrCacheRegistry(a, r)
	if err != nil {
		return nil, errors.Wrap(err, "cache registry")
	}

	return registrySpec, nil
}

func getOrCacheRegistry(a app.App, gh Registry) (*Spec, error) {
	// Check local disk cache.
	registrySpecFile := makePath(a, gh)
	registrySpec, exists, err := load(a, registrySpecFile)
	if err != nil {
		return nil, errors.Wrap(err, "load registry spec file")
	}

	if !exists {
		// If failed, use the protocol to try to retrieve app specification.
		registrySpec, err = gh.FetchRegistrySpec()
		if err != nil {
			return nil, err
		}

		registrySpecBytes, err := registrySpec.Marshal()
		if err != nil {
			return nil, err
		}

		// NOTE: We call mkdir after getting the registry spec, since a
		// network call might fail and leave this half-initialized empty
		// directory.
		registrySpecDir := filepath.Join(root(a), gh.RegistrySpecDir())
		err = a.Fs().MkdirAll(registrySpecDir, app.DefaultFolderPermissions)
		if err != nil {
			return nil, err
		}

		err = afero.WriteFile(a.Fs(), registrySpecFile, registrySpecBytes, app.DefaultFilePermissions)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	return registrySpec, nil
}

func load(a app.App, path string) (*Spec, bool, error) {
	exists, err := afero.Exists(a.Fs(), path)
	if err != nil {
		return nil, false, errors.Wrapf(err, "check if %q exists", path)
	}

	// NOTE: case where directory of the same name exists should be
	// fine, most filesystems allow you to have a directory and file of
	// the same name.
	if exists {
		isDir, err := afero.IsDir(a.Fs(), path)
		if err != nil {
			return nil, false, errors.Wrapf(err, "check if %q is a dir", path)
		}

		if !isDir {
			registrySpecBytes, err := afero.ReadFile(a.Fs(), path)
			if err != nil {
				return nil, false, err
			}

			registrySpec, err := Unmarshal(registrySpecBytes)
			if err != nil {
				return nil, false, err
			}
			return registrySpec, true, nil
		}
	}

	return nil, false, nil
}
