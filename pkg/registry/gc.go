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
	"io"
	"path/filepath"
	"strings"

	"github.com/ksonnet/ksonnet/pkg/pkg"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

// FsRemoveAller Subset of afero.Fs - just remove a directory
type FsRemoveAller interface {
	// RemoveAll removes a directory path and any children it contains. It
	// does not fail if the path does not exist (return nil).
	RemoveAll(path string) error
}

type vendorPathResolver interface {
	InstalledChecker
	VendorPath(pkg.Descriptor) (string, error)
}

// GarbageCollector removes vendored packages that are no longer needed
type GarbageCollector struct {
	pkgManager vendorPathResolver
	fs         FsRemoveAller
	root       string

	removeEmptyParentsFn func(path string, root string) error
}

// NewGarbageCollector constructs a GarbageCollector
func NewGarbageCollector(fs afero.Fs, pm vendorPathResolver, root string) GarbageCollector {
	return GarbageCollector{
		pkgManager: pm,
		fs:         fs,
		root:       root,

		removeEmptyParentsFn: func(path string, root string) error {
			return removeEmptyParents(fs, path, root)
		},
	}
}

// RemoveOrphans removes vendored packages that have been orphaned
func (gc GarbageCollector) RemoveOrphans(d pkg.Descriptor) error {
	log := log.WithField("action", "GarbageCollector.RemoveOrphans")
	installed, err := gc.pkgManager.IsInstalled(d)
	if err != nil {
		return errors.Wrapf(err, "checking installed status: %v", d)
	}
	// Only remove orphans
	if installed {
		return nil
	}

	path, err := gc.pkgManager.VendorPath(d)
	if err != nil {
		return errors.Wrapf(err, "resolving path for descriptor: %v", d)
	}

	if path == "" {
		return nil
	}

	if gc.fs == nil {
		return errors.New("nil fs")
	}
	log.Debugf("removing path %s", path)
	if err := gc.fs.RemoveAll(path); err != nil {
		return errors.Wrapf(err, "removing path %s for package %v", path, d)
	}

	if gc.removeEmptyParentsFn == nil {
		return nil
	}
	if err := gc.removeEmptyParentsFn(path, gc.root); err != nil {
		return errors.Wrapf(err, "removing empty parents path %s", path)
	}

	return nil
}

// Returns true if the specified directory is empty
func isDirEmpty(fs afero.Fs, path string) (bool, error) {
	if fs == nil {
		return false, errors.New("nil fs")
	}
	f, err := fs.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()
	_, err = f.Readdirnames(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err
}

// Remove empty directories, up to the provided root, exclusive
func removeEmptyParents(fs afero.Fs, path string, root string) error {
	if fs == nil {
		return errors.New("nil fs")
	}

	root = filepath.Clean(root)
	path = filepath.Clean(filepath.Dir(path))
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return errors.Wrapf(err, "%s not subpath of %s", path, root)
	}

	// while path is a subpath of root...
	for !strings.HasPrefix(rel, ".") {
		if fi, err := fs.Stat(path); err != nil || !fi.IsDir() {
			break
		}

		empty, err := isDirEmpty(fs, path)
		if err != nil {
			return errors.Wrapf(err, path)
		}
		if !empty {
			break
		}

		if err := fs.Remove(path); err != nil {
			return errors.Wrapf(err, path)
		}

		path = filepath.Dir(path)
		rel, err = filepath.Rel(root, path)
		if err != nil {
			return errors.Wrapf(err, "%s not subpath of %s", path, root)
		}
	}

	return nil
}
