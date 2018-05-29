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

package app

import (
	"io"

	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
)

// Encoder writes items to a serialized form.
type Encoder interface {
	// Encode writes an item to a stream. Implementations may return errors
	// if the data to be encoded is invalid.
	Encode(i interface{}, w io.Writer) error
}

var (
	defaultYAMLEncoder = &YAMLEncoder{}
)

// YAMLEncoder write items to a serialized form in YAML format.
type YAMLEncoder struct{}

// Encode encodes data in yaml format.
func (e *YAMLEncoder) Encode(i interface{}, w io.Writer) error {
	b, err := yaml.Marshal(i)
	if err != nil {
		return errors.Wrap(err, "encoding data")
	}

	_, err = w.Write(b)
	return err
}
