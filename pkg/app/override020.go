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

// Override020 defines overrides to ksonnet project configurations.
type Override020 struct {
	Kind         string                `json:"kind"`
	APIVersion   string                `json:"apiVersion"`
	Environments EnvironmentConfigs020 `json:"environments,omitempty"`
	Registries   RegistryConfigs020    `json:"registries,omitempty"`
}

type override020 Override020

// UnmarshalJSON implements the json.Unmarshaler interface.
func (o *Override020) UnmarshalJSON(b []byte) error {
	var r override020
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
		r.Environments = EnvironmentConfigs020{}
	}
	if r.Registries == nil {
		r.Registries = RegistryConfigs020{}
	}

	*o = Override020(r)
	return nil
}
