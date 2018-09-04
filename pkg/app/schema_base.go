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

	"github.com/blang/semver"
	"github.com/pkg/errors"
)

type specBase struct {
	APIVersion semver.Version `json:"apiVersion,omitempty"`
	Kind       string         `json:"kind,omitempty"`
}

type specBaseRaw struct {
	APIVersion string `json:"apiVersion,omitempty"`
	Kind       string `json:"kind,omitempty"`
}

func (s *specBase) UnmarshalJSON(b []byte) error {
	var raw specBaseRaw
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}

	version, err := semver.Parse(raw.APIVersion)
	if err != nil {
		return errors.Wrapf(err, "parsing semver: %s", raw.APIVersion)
	}

	s.APIVersion = version
	s.Kind = raw.Kind
	return nil
}
