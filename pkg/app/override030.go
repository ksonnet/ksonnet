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
	"encoding/json"

	"github.com/pkg/errors"
)

// Override030 defines overrides to ksonnet project configurations.
type Override030 struct {
	Kind         string                `json:"kind"`
	APIVersion   string                `json:"apiVersion"`
	Environments EnvironmentConfigs030 `json:"environments,omitempty"`
	Registries   RegistryConfigs030    `json:"registries,omitempty"`
}

// Validate validates an Override.
func (o *Override030) Validate() error {
	if o.Kind != overrideKind {
		return errors.Errorf("app override has unexpected kind")
	}

	if o.APIVersion != overrideVersion {
		return errors.Errorf("app override has unexpected apiVersion")
	}

	return nil
}

// IsDefined returns true if the override has environments or registries defined.
func (o *Override030) IsDefined() bool {
	return o != nil && (len(o.Environments) > 0 || len(o.Registries) > 0)
}

type override030 Override030

// UnmarshalJSON implements the json.Unmarshaler interface.
func (o *Override030) UnmarshalJSON(b []byte) error {
	var r override030
	if err := json.Unmarshal(b, &r); err != nil {
		return err
	}

	if r.Kind != overrideKind {
		return errors.Errorf("app override has unexpected kind")
	}

	if r.APIVersion != overrideVersion {
		return errors.Errorf("app override has unexpected apiVersion")
	}

	if r.Environments == nil {
		r.Environments = EnvironmentConfigs030{}
	}
	if r.Registries == nil {
		r.Registries = RegistryConfigs030{}
	}

	*o = Override030(r)
	return nil
}
