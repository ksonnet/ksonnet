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
	rice "github.com/GeertJohan/go.rice"
	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
)

var (
	// DefaultBuilder is a builder that will build a prototype from a source string.
	DefaultBuilder = JsonnetParse
)

// Builder builds a prototype from a source string.
type Builder func(source string) (*Prototype, error)

// Unmarshal takes the bytes of a JSON-encoded prototype specification, and
// de-serializes them to a `SpecificationSchema`.
func Unmarshal(bytes []byte) (*Prototype, error) {
	var p Prototype
	err := yaml.Unmarshal(bytes, &p)
	if err != nil {
		return nil, err
	}

	if err = p.validate(); err != nil {
		return nil, err
	}

	return &p, nil
}

// SearchOptions represents the type of prototype search to execute on an
// `Index`.
type SearchOptions int

const (
	// Prefix represents a search over prototype name prefixes.
	Prefix SearchOptions = iota

	// Suffix represents a search over prototype name suffixes.
	Suffix

	// Substring represents a search over substrings of prototype names.
	Substring
)

// Index represents a queryable index of prototype specifications.
type Index interface {
	List() (Prototypes, error)
	SearchNames(query string, opts SearchOptions) (Prototypes, error)
}

// NewIndex constructs an index of prototype specifications from a list.
func NewIndex(prototypes []*Prototype, builder Builder) (Index, error) {
	idx := map[string]*Prototype{}

	systemBox, err := rice.FindBox("system")
	if err != nil {
		return nil, err
	}

	dp, err := systemPrototypes(systemBox, builder)
	if err != nil {
		return nil, errors.Wrap(err, "unable to load default prototypes")
	}

	for _, p := range dp {
		idx[p.Name] = p
	}

	for _, p := range prototypes {
		idx[p.Name] = p
	}

	return &index{
		prototypes: idx,
	}, nil
}
