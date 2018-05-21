// Copyright 2018 The kubecfg authors
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

package prototype

import (
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

//go:generate rice embed-go

type container interface {
	String(path string) (string, error)
	Walk(path string, fn filepath.WalkFunc) error
}

func systemPrototypes(c container, pb Builder) ([]*Prototype, error) {
	var prototypes []*Prototype

	err := c.Walk("", func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if fi.IsDir() {
			return nil
		}

		source, err := c.String(path)
		if err != nil {
			return errors.Wrapf(err, "%q does not exist", path)
		}

		prototype, err := pb(source)
		if err != nil {
			return errors.WithStack(err)
		}

		prototypes = append(prototypes, prototype)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return prototypes, nil
}
