// Copyright 2017 The kubecfg authors
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
	"encoding/json"
	"fmt"

	"github.com/blang/semver"
	"github.com/ghodss/yaml"
	"github.com/ksonnet/ksonnet/pkg/app"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
)

const (
	// DefaultAPIVersion is the default version of the registry API.
	DefaultAPIVersion = "0.1.0"
	// DefaultKind is the default kind of the registry API.
	DefaultKind = "ksonnet.io/registry"
)

// Spec describes how a registry is stored.
type Spec struct {
	APIVersion string         `json:"apiVersion"`
	Kind       string         `json:"kind"`
	Version    string         `json:"version"`
	Libraries  LibraryConfigs `json:"libraries"`
}

// specDeprecated is the previous registry specification
type specDeprecated struct {
	APIVersion string              `json:"apiVersion"`
	Kind       string              `json:"kind"`
	GitVersion *app.GitVersionSpec `json:"gitVersion"`
	Libraries  LibraryConfigs      `json:"libraries"`
}

// spec is an alias that allows us to leverage default JSON decoding
//  in our custom UnmarshalJSON handler without triggering infinite recursion.
type spec Spec

// UnmarshalJSON implements the json.Unmarshaler interface.
// We implement some compatibility conversions.
func (s *Spec) UnmarshalJSON(b []byte) error {
	var newSpec spec

	if err := json.Unmarshal(b, &newSpec); err != nil {
		return err
	}
	*s = Spec(newSpec)

	// Check if there's any need for conversions
	if s.Version != "" {
		return nil
	}

	// Try to convert deprecated fields
	var oldStyle specDeprecated
	if err := json.Unmarshal(b, &oldStyle); err != nil {
		// This is best-effort, not an error
		return nil
	}
	if oldStyle.GitVersion != nil {
		s.Version = oldStyle.GitVersion.CommitSHA
	}

	return nil
}

// Unmarshal unmarshals bytes to a Spec.
func Unmarshal(bytes []byte) (*Spec, error) {
	schema := Spec{}
	err := yaml.Unmarshal(bytes, &schema)
	if err != nil {
		return nil, err
	}

	if err = schema.validate(); err != nil {
		return nil, err
	}

	return &schema, nil
}

// Marshal marshals a Spec to YAML.
func (s *Spec) Marshal() ([]byte, error) {
	return yaml.Marshal(s)
}

func (s *Spec) validate() error {
	// Originally, the default value for `apiVersion` was `0.1`. This is not a
	// valid semver, so before we do anything, we need to convert it to one.
	if s.APIVersion == "0.1" {
		s.APIVersion = "0.1.0"
	}

	compatVer, _ := semver.Make(DefaultAPIVersion)
	ver, err := semver.Make(s.APIVersion)
	if err != nil {
		return errors.Wrap(err, "Failed to parse version in app spec")
	} else if compatVer.Compare(ver) != 0 {
		return fmt.Errorf(
			"Registry uses unsupported spec version '%s' (this client only supports %s)",
			s.APIVersion,
			DefaultAPIVersion)
	}

	return nil
}

// load loads a registry spec from disk.
// Returns the parsed spec, bool if it existed, and optional error.
func load(a app.App, path string) (*Spec, bool, error) {
	exists, err := afero.Exists(a.Fs(), path)
	if err != nil {
		return nil, false, errors.Wrapf(err, "check if %q exists", path)
	}

	// NOTE: case where directory of the same name exists should be
	// fine, most filesystems allow you to have a directory and file of
	// the same name.
	if !exists {
		return nil, false, nil
	}

	isDir, err := afero.IsDir(a.Fs(), path)
	if err != nil {
		return nil, false, errors.Wrapf(err, "check if %q is a dir", path)
	}

	if isDir {
		return nil, false, nil
	}

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

// Specs is a slice of *Spec.
type Specs []*Spec

// LibaryConfig is library reference.
type LibaryConfig struct {
	Version string `json:"version"`
	Path    string `json:"path"`
}

// LibraryConfigs maps LibraryConfigs to a name.
type LibraryConfigs map[string]*LibaryConfig
