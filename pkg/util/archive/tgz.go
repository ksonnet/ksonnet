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

package archive

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"io"
)

// Tgz handles gzip'd tar archives.
type Tgz struct {
}

// UnTgz gunzips and un-tars the contents of a reader. Then handler
// will be called for every file in the archive.
func (t *Tgz) Unarchive(r io.Reader, handler FileHandler) error {
	if r == nil {
		return errors.New("gzip reader is nil")
	}

	gzReader, err := gzip.NewReader(r)
	if err != nil {
		return err
	}

	tarReader := tar.NewReader(gzReader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}

		name := header.Name

		switch header.Typeflag {
		case 0:
			tf := &File{
				Name:   name,
				Reader: tarReader,
			}

			if err = handler(tf); err != nil {
				return err
			}
		}
	}
	return nil
}
