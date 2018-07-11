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

package cluster

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"

	"github.com/ksonnet/ksonnet/pkg/util/serial"
)

// managedAnnotation is the contents of of the ksonnet.io/managed annotation.
type managedAnnotation struct {
	Pristine string `json:"pristine,omitempty"`
}

// Marshal marshals this object to JSON.
func (mm *managedAnnotation) Marshal() ([]byte, error) {
	return json.Marshal(mm)
}

// Encode encodes a pristine copy of the object.
func (mm *managedAnnotation) Encode(m map[string]interface{}) error {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)

	actions := []serial.Action{
		func() error { return json.NewEncoder(gz).Encode(m) },
		gz.Flush,
		gz.Close,
	}

	if err := serial.RunActions(actions...); err != nil {
		return err
	}

	mm.Pristine = base64.StdEncoding.EncodeToString(buf.Bytes())
	return nil
}

// Decode decodes a pristine copy of the object.
func (mm *managedAnnotation) Decode() (map[string]interface{}, error) {
	b, err := base64.StdEncoding.DecodeString(mm.Pristine)
	if err != nil {
		return nil, err
	}

	r := bytes.NewReader(b)

	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}
	defer zr.Close()

	var m map[string]interface{}
	if err := json.NewDecoder(zr).Decode(&m); err != nil {
		return nil, err
	}

	return m, nil
}
