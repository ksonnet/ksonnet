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

package kubecfg

import (
	"github.com/ksonnet/ksonnet/metadata"
)

type EnvAddCmd struct {
	name      string
	server    string
	namespace string
	spec      string

	manager metadata.Manager
}

func NewEnvAddCmd(name, server, namespace, specFlag string, manager metadata.Manager) (*EnvAddCmd, error) {
	return &EnvAddCmd{name: name, server: server, namespace: namespace, spec: specFlag, manager: manager}, nil
}

func (c *EnvAddCmd) Run() error {
	return c.manager.CreateEnvironment(c.name, c.server, c.namespace, c.spec)
}

// ==================================================================

type EnvRmCmd struct {
	name string

	manager metadata.Manager
}

func NewEnvRmCmd(name string, manager metadata.Manager) (*EnvRmCmd, error) {
	return &EnvRmCmd{name: name, manager: manager}, nil
}

func (c *EnvRmCmd) Run() error {
	return c.manager.DeleteEnvironment(c.name)
}
