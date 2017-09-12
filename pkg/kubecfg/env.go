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
	"sort"
	"strings"

	"github.com/ksonnet/kubecfg/metadata"
)

type EnvAddCmd struct {
	name string
	uri  string

	rootPath metadata.AbsPath
	spec     metadata.ClusterSpec
}

func NewEnvAddCmd(name, uri, specFlag string, rootPath metadata.AbsPath) (*EnvAddCmd, error) {
	spec, err := metadata.ParseClusterSpec(specFlag)
	if err != nil {
		return nil, err
	}

	return &EnvAddCmd{name: name, uri: uri, spec: spec, rootPath: rootPath}, nil
}

func (c *EnvAddCmd) Run() error {
	manager, err := metadata.Find(c.rootPath)
	if err != nil {
		return err
	}

	extensionsLibData, k8sLibData, err := manager.GenerateKsonnetLibData(c.spec)
	if err != nil {
		return err
	}

	return manager.CreateEnvironment(c.name, c.uri, c.spec, extensionsLibData, k8sLibData)
}

// ==================================================================

type EnvListCmd struct {
	rootPath metadata.AbsPath
}

func NewEnvListCmd(rootPath metadata.AbsPath) (*EnvListCmd, error) {
	return &EnvListCmd{rootPath: rootPath}, nil
}

func (c *EnvListCmd) Run() (string, error) {
	manager, err := metadata.Find(c.rootPath)
	if err != nil {
		return "", err
	}

	envs, err := manager.GetEnvironments()
	if err != nil {
		return "", err
	}

	// Sort environments by ascending alphabetical name
	sort.Slice(envs, func(i, j int) bool { return envs[i].Name < envs[j].Name })

	// Format each environment information for pretty printing.
	// Each environment should be outputted like the following:
	//
	//  us-west/dev     localhost:8080
	//  us-west/staging http://example.com
	//
	// To accomplish this, need to find the longest env name for proper padding.

	maxNameLen := 0
	for _, env := range envs {
		if l := len(env.Name); l > maxNameLen {
			maxNameLen = l
		}
	}

	lines := []string{}
	for _, env := range envs {
		nameSpacing := strings.Repeat(" ", maxNameLen-len(env.Name)+1)
		lines = append(lines, env.Name+nameSpacing+env.URI+"\n")
	}

	return strings.Join(lines, ""), nil
}
