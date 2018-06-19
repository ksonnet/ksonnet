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
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/afero"
)

const (
	// overrideKind is the override resource type.
	overrideKind = "ksonnet.io/app-override"
	// overrideVersion is the version of the override resource.
	overrideVersion = "0.1.0"
)

// Override defines overrides to ksonnet project configurations.
type Override struct {
	Kind         string           `json:"kind"`
	APIVersion   string           `json:"apiVersion"`
	Environments EnvironmentSpecs `json:"environments,omitempty"`
	Registries   RegistryConfigs  `json:"registries,omitempty"`
}

// Validate validates an Override.
func (o *Override) Validate() error {
	if o.Kind != overrideKind {
		return errors.Errorf("app override has unexpected kind")
	}

	if o.APIVersion != overrideVersion {
		return errors.Errorf("app override has unexpected apiVersion")
	}

	return nil
}

// IsDefined returns true if the override has environments or registries defined.
func (o *Override) IsDefined() bool {
	return len(o.Environments) > 0 || len(o.Registries) > 0
}

// SaveOverride saves the override to the filesystem.
func SaveOverride(encoder Encoder, fs afero.Fs, root string, o *Override) error {
	if o == nil {
		return errors.New("override was nil")
	}

	o.APIVersion = overrideVersion
	o.Kind = overrideKind

	f, err := fs.OpenFile(overridePath(root), os.O_WRONLY|os.O_CREATE, DefaultFilePermissions)
	if err != nil {
		return err
	}

	if err := encoder.Encode(o, f); err != nil {
		return errors.Wrap(err, "encoding override")
	}

	return nil
}
