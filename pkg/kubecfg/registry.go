// Copyright 2017 The ksonnet authors
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

package kubecfg

// RegistryAddCmd contains the metadata needed to create a registry.
type RegistryAddCmd struct {
	name     string
	protocol string
	uri      string
	version  string
}

// NewRegistryAddCmd initializes a RegistryAddCmd.
func NewRegistryAddCmd(name, protocol, uri, version string) *RegistryAddCmd {
	if version == "" {
		version = "latest"
	}

	return &RegistryAddCmd{name: name, protocol: protocol, uri: uri, version: version}
}

// Run adds the registry to the ksonnet project.
func (c *RegistryAddCmd) Run() error {
	manager, err := manager()
	if err != nil {
		return err
	}

	_, err = manager.AddRegistry(c.name, c.protocol, c.uri, c.version)
	return err
}
