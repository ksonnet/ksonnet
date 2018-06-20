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

package io

import (
	"io"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/spf13/afero"
)

// CopyRecursive copies the contents of the src directory to the dst directory, recursively.
// Equivilant to to `cp -R src dst`
func CopyRecursive(fs afero.Fs, dst string, src string, fileMode os.FileMode, dirMode os.FileMode) error {
	if fs == nil {
		return errors.Errorf("nil filesystem interface")
	}

	return afero.Walk(fs, src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return errors.Wrapf(err, "extracting relative path: %v, %v", src, path)
		}
		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			err := fs.MkdirAll(dstPath, dirMode)
			if err != nil {
				return errors.Wrapf(err, "creating %v", dstPath)
			}
			return nil
		}

		out, err := fs.Create(dstPath)
		if err != nil {
			return errors.Wrapf(err, "creating %v", dstPath)
		}
		defer out.Close()

		in, err := fs.Open(path)
		if err != nil {
			return errors.Wrapf(err, "opening %v", path)
		}
		defer in.Close()
		if _, err = io.Copy(out, in); err != nil {
			return errors.Wrapf(err, "copying contents of %v", path)
		}
		return nil
	})
}
