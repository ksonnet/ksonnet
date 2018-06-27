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

package actions

import (
	"net/url"
	"strings"

	"github.com/ksonnet/ksonnet/pkg/app"
	"github.com/ksonnet/ksonnet/pkg/registry"
	"github.com/pkg/errors"
)

// RunRegistryAdd runs `registry add`
func RunRegistryAdd(m map[string]interface{}) error {
	ra, err := NewRegistryAdd(m)
	if err != nil {
		return err
	}

	return ra.Run()
}

// RegistryAdd adds a registry.
type RegistryAdd struct {
	app           app.App
	name          string
	uri           string
	isOverride    bool
	registryAddFn func(a app.App, protocol registry.Protocol, name string, uri string, isOverride bool) (*registry.Spec, error)
}

// NewRegistryAdd creates an instance of RegistryAdd.
func NewRegistryAdd(m map[string]interface{}) (*RegistryAdd, error) {
	ol := newOptionLoader(m)

	ra := &RegistryAdd{
		app:        ol.LoadApp(),
		name:       ol.LoadString(OptionName),
		uri:        ol.LoadString(OptionURI),
		isOverride: ol.LoadBool(OptionOverride),

		registryAddFn: registry.Add,
	}

	if ol.err != nil {
		return nil, ol.err
	}

	return ra, nil
}

// Run adds a registry.
func (ra *RegistryAdd) Run() error {
	rd, err := ra.protocol()
	if err != nil {
		return errors.Wrap(err, "detect registry protocol")
	}

	_, err = ra.registryAddFn(ra.app, rd.Protocol, ra.name, rd.URI, ra.isOverride)
	return err
}

type registryDetails struct {
	URI      string
	Protocol registry.Protocol
}

func (ra *RegistryAdd) protocol() (registryDetails, error) {
	if ra.isGitHub() {
		rd := registryDetails{
			URI:      ra.uri,
			Protocol: registry.ProtocolGitHub,
		}

		return rd, nil
	}

	if strings.HasPrefix(ra.uri, "file://") {
		u, err := url.Parse(ra.uri)
		if err != nil {
			return registryDetails{}, err
		}

		rd := registryDetails{
			URI:      u.Path,
			Protocol: registry.ProtocolFilesystem,
		}

		return rd, nil
	}

	if strings.HasPrefix(ra.uri, "/") || strings.HasPrefix(ra.uri, ".") {
		rd := registryDetails{
			URI:      ra.uri,
			Protocol: registry.ProtocolFilesystem,
		}

		return rd, nil
	}

	_, err := url.Parse(ra.uri)
	if err == nil {
		// assuming uri is a helm repository URL
		rd := registryDetails{
			URI:      ra.uri,
			Protocol: registry.ProtocolHelm,
		}

		return rd, nil
	}

	return registryDetails{}, errors.Errorf("could not detect registry type for %s", ra.uri)
}

func (ra *RegistryAdd) isGitHub() bool {
	return strings.HasPrefix(ra.uri, "github.com") ||
		strings.HasPrefix(ra.uri, "https://github.com")
}
