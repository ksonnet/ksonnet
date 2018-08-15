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
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/ksonnet/ksonnet/pkg/app"
	"github.com/ksonnet/ksonnet/pkg/pkg"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

// CacheDependency vendors registry dependencies.
// TODO: create unit tests for this once mocks for this package are
// worked out.
func CacheDependency(a app.App, checker InstalledChecker, d pkg.Descriptor, customName string, force bool, httpClient *http.Client) (*app.LibraryConfig, error) {
	logger := log.WithFields(log.Fields{
		"action":      "registry.CacheDependency",
		"part":        d.Name,
		"registry":    d.Registry,
		"version":     d.Version,
		"custom-name": customName,
	})

	if a == nil {
		return nil, errors.Errorf("nil receiver")
	}
	if checker == nil {
		return nil, errors.Errorf("nil installation checker")
	}

	logger.Debug("caching dependency")

	registries, err := a.Registries()
	if err != nil {
		return nil, err
	}

	regRefSpec, exists := registries[d.Registry]
	if !exists {
		return nil, fmt.Errorf("registry '%s' does not exist", d.Registry)
	}

	r, err := Locate(a, regRefSpec, httpClient)
	if err != nil {
		return nil, err
	}

	// Get all directories and files first, then write to disk. This
	// protects us from failing with a half-cached dependency because of
	// a network failure.
	directories := []string{}
	files := map[string][]byte{}
	_, libRef, err := r.ResolveLibrary(
		d.Name,
		customName,
		d.Version,
		func(relPath string, contents []byte) error {
			files[relPath] = contents
			return nil
		},
		func(relPath string) error {
			return nil
		})
	if err != nil {
		return nil, errors.Wrap(err, "resolve registry library")
	}

	// Make triple-sure the library references the correct registry, as it is known in this app.
	libRef.Registry = d.Registry

	// Check whether this library version is already installed
	var qualified = d
	qualified.Version = libRef.Version
	ok, err := checker.IsInstalled(qualified)
	if err != nil {
		return nil, errors.Wrapf(err, "checking package installed status")
	}
	if ok && !force {
		return nil, errors.Errorf("package '%s/%s@%s' already exists.",
			libRef.Registry, libRef.Name, libRef.Version)
	}

	// Add library to app specification, but wait to write it out until
	// the end, in case one of the network calls fails.
	log.Infof("Retrieved %d files", len(files))

	for _, dir := range directories {
		if err = a.Fs().MkdirAll(dir, app.DefaultFolderPermissions); err != nil {
			return nil, errors.Wrap(err, "unable to create directory")
		}
	}

	vendorRoot := a.VendorPath()
	for path, content := range files {
		vendoredPath := versionAndVendorRelPath(libRef, vendorRoot, path)
		if vendoredPath == "" {
			log.Warnf("problem vendoring file: %v", path)
			continue
		}
		dir := filepath.Dir(filepath.FromSlash(vendoredPath))

		log.Debugf("onFile: vendoring file to path: %v", vendoredPath)
		if err = a.Fs().MkdirAll(dir, app.DefaultFolderPermissions); err != nil {
			return nil, errors.Wrap(err, "unable to create directory")
		}

		if err = afero.WriteFile(a.Fs(), vendoredPath, content, app.DefaultFilePermissions); err != nil {
			return nil, errors.Wrap(err, "unable to create file")
		}
	}

	return libRef, nil
}

// Convert a relative path like `mysql/parts.yaml` to a versioned, vendored path,
// like `<app_root>/vendor/<registry>/mysql@0011223344/parts.yaml`
// Assumption: paths are relative to the registry root (not repo root!)
func versionAndVendorRelPath(lib *app.LibraryConfig, vendorRoot string, relPath string) string {
	if lib == nil {
		return ""
	}

	// Version the path
	var versionedPath string
	if lib.Version != "" {
		//filepath.ToSlash()
		parts := strings.SplitN(filepath.ToSlash(relPath), "/", -1)
		if parts[0] == lib.Name {
			parts[0] = fmt.Sprintf("%s@%s", lib.Name, lib.Version)
		}
		versionedPath = filepath.FromSlash(strings.Join(parts, "/"))

		// oldPrefix := filepath.Join(lib.Registry, lib.Name)
		// newPrefix := fmt.Sprintf("%s@%s", lib.Name, lib.Version)
		// versionedPath = strings.Replace(relPath, oldPrefix, newPrefix, 1)
	} else {
		// For unversioned packages, use path as-is
		versionedPath = relPath
	}

	vendorFilePath := filepath.Join(vendorRoot, lib.Registry, versionedPath)
	return vendorFilePath
}
