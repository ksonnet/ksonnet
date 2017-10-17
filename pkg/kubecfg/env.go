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
)

type EnvAddCmd struct {
	name      string
	uri       string
	namespace string

	spec    metadata.ClusterSpec
	manager metadata.Manager
}

func NewEnvAddCmd(name, uri, namespace, specFlag string, manager metadata.Manager) (*EnvAddCmd, error) {
	spec, err := metadata.ParseClusterSpec(specFlag)
	if err != nil {
		return nil, err
	}
	log.Debugf("Generating ksonnetLib data with spec: %s", specFlag)

	return &EnvAddCmd{name: name, uri: uri, namespace: namespace, spec: spec, manager: manager}, nil
}

func (c *EnvAddCmd) Run() error {
	return c.manager.CreateEnvironment(c.name, c.uri, c.namespace, c.spec)
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
		uriHeader       = "URI"
	)

	envs, err := c.manager.GetEnvironments()
	if err != nil {
		return err
	}

	// Sort environments by ascending alphabetical name
	sort.Slice(envs, func(i, j int) bool { return envs[i].Name < envs[j].Name })

	// Format each environment information for pretty printing.
	// Each environment should be outputted like the following:
	//
	//   NAME            NAMESPACE URI
	//   minikube        dev       localhost:8080
	//   us-west/staging staging   http://example.com
	//
	// To accomplish this, need to find the longest env name and the longest
	// env namespace for proper padding.

	maxNameLen := len(nameHeader)
	for _, env := range envs {
		if l := len(env.Name); l > maxNameLen {
			maxNameLen = l
		}
	}

	maxNamespaceLen := len(namespaceHeader) + maxNameLen + 1
	for _, env := range envs {
		if l := len(env.Namespace) + maxNameLen + 1; l > maxNamespaceLen {
			maxNamespaceLen = l
		}
	}

	lines := []string{}

	headerNameSpacing := strings.Repeat(" ", maxNameLen-len(nameHeader)+1)
	headerNamespaceSpacing := strings.Repeat(" ", maxNamespaceLen-maxNameLen-len(namespaceHeader))
	lines = append(lines, nameHeader+headerNameSpacing+namespaceHeader+headerNamespaceSpacing+uriHeader+"\n")

	for _, env := range envs {
		nameSpacing := strings.Repeat(" ", maxNameLen-len(env.Name)+1)
		namespaceSpacing := strings.Repeat(" ", maxNamespaceLen-maxNameLen-len(env.Namespace))
		lines = append(lines, env.Name+nameSpacing+env.Namespace+namespaceSpacing+env.URI+"\n")
	}

	formattedEnvsList := strings.Join(lines, "")

	_, err = fmt.Fprint(out, formattedEnvsList)
	return err
}

// ==================================================================

type EnvSetCmd struct {
	name string

	desiredName      string
	desiredURI       string
	desiredNamespace string

	manager metadata.Manager
}

func NewEnvSetCmd(name, desiredName, desiredURI, desiredNamespace string, manager metadata.Manager) (*EnvSetCmd, error) {
	return &EnvSetCmd{name: name, desiredName: desiredName, desiredURI: desiredURI, desiredNamespace: desiredNamespace,
		manager: manager}, nil
}

func (c *EnvSetCmd) Run() error {
	desired := metadata.Environment{Name: c.desiredName, URI: c.desiredURI, Namespace: c.desiredNamespace}
	return c.manager.SetEnvironment(c.name, &desired)
}
