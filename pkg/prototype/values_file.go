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

package prototype

import (
	"io"
	"io/ioutil"
	"sort"
	"strings"

	"github.com/ksonnet/ksonnet/pkg/util/jsonnet"
	"github.com/pkg/errors"
)

// ValuesFile represents a prototype values file.
type ValuesFile struct {
	src string
}

// NewValuesFile creates an instance of ValuesFile.
func NewValuesFile(src string) *ValuesFile {
	return &ValuesFile{
		src: src,
	}
}

// Keys returns the keys in the values file.
func (vf *ValuesFile) Keys() ([]string, error) {
	obj, err := jsonnet.Parse("values-file", vf.src)
	if err != nil {
		return nil, errors.Wrap(err, "parsing values-file")
	}

	var keys []string

	for i := range obj.Fields {
		f := obj.Fields[i]

		id, err := jsonnet.FieldID(f)
		if err != nil {
			return nil, errors.Wrap(err, "finding field in jsonnet object")
		}

		keys = append(keys, id)
	}

	sort.Strings(keys)

	return keys, nil
}

// Get gets a value from the values file by key.
func (vf *ValuesFile) Get(k string) (string, error) {
	vm := jsonnet.NewVM()
	vm.TLACode("object", vf.src)
	vm.TLAVar("key", k)

	v, err := vm.EvaluateSnippet("getValue", snippetFieldValue)
	if err != nil {
		return "", errors.Errorf("key %q was not found", k)
	}

	return strings.TrimSpace(v), nil
}

var snippetFieldValue = `
function(object, key)
	object[key]
`

// ReadValues reads a values file from a reader.
func ReadValues(r io.Reader) (*ValuesFile, error) {
	if r == nil {
		return nil, errors.Errorf("reader is nil")
	}

	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, errors.Wrap(err, "reading values")
	}

	vm := jsonnet.NewVM()

	evaluated, err := vm.EvaluateSnippet("prototype-values", string(data))
	if err != nil {
		return nil, errors.Wrap(err, "evaluating values with jsonnet")
	}

	return &ValuesFile{
		src: evaluated,
	}, nil
}
