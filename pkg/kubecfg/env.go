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
	"fmt"
	"io"
	"sort"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/ksonnet/ksonnet/metadata"
	"github.com/ksonnet/ksonnet/utils"
)

type EnvAddCmd struct {
	name      string
	server    string
	namespace string

	spec    metadata.ClusterSpec
	manager metadata.Manager
}

func NewEnvAddCmd(name, server, namespace, specFlag string, manager metadata.Manager) (*EnvAddCmd, error) {
	spec, err := metadata.ParseClusterSpec(specFlag)
	if err != nil {
		return nil, err
	}
	log.Debugf("Generating ksonnetLib data with spec: %s", specFlag)

	return &EnvAddCmd{name: name, server: server, namespace: namespace, spec: spec, manager: manager}, nil
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

// ==================================================================

type EnvListCmd struct {
	manager metadata.Manager
}

func NewEnvListCmd(manager metadata.Manager) (*EnvListCmd, error) {
	return &EnvListCmd{manager: manager}, nil
}

func (c *EnvListCmd) Run(out io.Writer) error {
	const (
		nameHeader      = "NAME"
		namespaceHeader = "NAMESPACE"
		serverHeader    = "SERVER"
	)

	envs, err := c.manager.GetEnvironments()
	if err != nil {
		return err
	}

	// Sort environments by ascending alphabetical name
	sort.Slice(envs, func(i, j int) bool { return envs[i].Name < envs[j].Name })

	rows := [][]string{
		[]string{nameHeader, namespaceHeader, serverHeader},
		[]string{
			strings.Repeat("=", len(nameHeader)),
			strings.Repeat("=", len(namespaceHeader)),
			strings.Repeat("=", len(serverHeader))},
	}
	for _, env := range envs {
		rows = append(rows, []string{env.Name, env.Namespace, env.Server})
	}

	formattedEnvsList, err := utils.PadRows(rows)
	if err != nil {
		return err
	}

	_, err = fmt.Fprint(out, formattedEnvsList)
	return err
}

// ==================================================================

type EnvSetCmd struct {
	name string

	desiredName      string
	desiredServer    string
	desiredNamespace string

	manager metadata.Manager
}

func NewEnvSetCmd(name, desiredName, desiredServer, desiredNamespace string, manager metadata.Manager) (*EnvSetCmd, error) {
	return &EnvSetCmd{name: name, desiredName: desiredName, desiredServer: desiredServer, desiredNamespace: desiredNamespace,
		manager: manager}, nil
}

func (c *EnvSetCmd) Run() error {
	desired := metadata.Environment{Name: c.desiredName, Server: c.desiredServer, Namespace: c.desiredNamespace}
	return c.manager.SetEnvironment(c.name, &desired)
}
