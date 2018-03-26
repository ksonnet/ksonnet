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

package metadata

import (
	"bytes"
	"fmt"
	"runtime/debug"

	"github.com/ksonnet/ksonnet/env"
	"github.com/ksonnet/ksonnet/metadata/lib"
	str "github.com/ksonnet/ksonnet/strings"
	"github.com/pkg/errors"

	log "github.com/sirupsen/logrus"

	param "github.com/ksonnet/ksonnet/metadata/params"
)

const (
	defaultEnvName = "default"

	// primary environment files
	envFileName    = "main.jsonnet"
	paramsFileName = "params.libsonnet"
)

var (
	// envCreate is a function which creates environments.
	envCreate = env.Create
)

func (m *manager) CreateEnvironment(name, server, namespace, k8sSpecFlag string) error {
	debug.PrintStack()
	return errors.Errorf("deprecated")
	// a, err := m.App()
	// if err != nil {
	// 	return err
	// }

	// config := env.CreateConfig{
	// 	App:         a,
	// 	Destination: env.NewDestination(server, namespace),
	// 	K8sSpecFlag: k8sSpecFlag,
	// 	Name:        name,

	// 	OverrideData: m.generateOverrideData(),
	// 	ParamsData:   m.generateParamsData(),
	// }

	// return envCreate(config)

}

func (m *manager) DeleteEnvironment(name string) error {
	a, err := m.App()
	if err != nil {
		return err
	}

	// TODO: move this to actions
	return env.Delete(a, name, false)
}

func (m *manager) GetEnvironments() (map[string]env.Env, error) {
	a, err := m.App()
	if err != nil {
		return nil, err
	}

	log.Debug("Retrieving all environments")
	return env.List(a)
}

func (m *manager) GetEnvironment(name string) (*env.Env, error) {
	a, err := m.App()
	if err != nil {
		return nil, err
	}

	return env.Retrieve(a, name)
}

func (m *manager) SetEnvironment(from, to string) error {
	a, err := m.App()
	if err != nil {
		return err
	}

	// TODO: move this to an action
	return env.Rename(a, from, to, false)
}

func (m *manager) GetEnvironmentParams(name, nsName string) (map[string]param.Params, error) {
	a, err := m.App()
	if err != nil {
		return nil, err
	}

	config := env.GetParamsConfig{
		App: a,
	}

	return env.GetParams(name, nsName, config)
}

func (m *manager) SetEnvironmentParams(envName, component string, params param.Params) error {
	a, err := m.App()
	if err != nil {
		return err
	}

	config := env.SetParamsConfig{
		App: a,
	}

	return env.SetParams(envName, component, params, config)
}

func (m *manager) EnvPaths(env string) (libPath, mainPath, paramsPath string, err error) {
	mainPath, paramsPath = m.makeEnvPaths(env)
	libPath, err = m.getLibPath(env)
	return
}

func (m *manager) makeEnvPaths(env string) (mainPath, paramsPath string) {
	envPath := str.AppendToPath(m.environmentsPath, env)

	// main.jsonnet file
	mainPath = str.AppendToPath(envPath, envFileName)
	// params.libsonnet file
	paramsPath = str.AppendToPath(envPath, componentParamsFile)

	return
}

func (m *manager) getLibPath(envName string) (string, error) {
	a, err := m.App()
	if err != nil {
		return "", err
	}

	return a.LibPath(envName)
}

func (m *manager) generateOverrideData() []byte {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("local base = import \"%s\";\n", baseLibsonnetFile))
	buf.WriteString(fmt.Sprintf("local k = import \"%s\";\n\n", lib.ExtensionsLibFilename))
	buf.WriteString("base + {\n")
	buf.WriteString("  // Insert user-specified overrides here. For example if a component is named \"nginx-deployment\", you might have something like:\n")
	buf.WriteString("  //   \"nginx-deployment\"+: k.deployment.mixin.metadata.labels({foo: \"bar\"})\n")
	buf.WriteString("}\n")
	return buf.Bytes()
}

func (m *manager) generateParamsData() []byte {
	const (
		relComponentParamsPath = "../../" + componentsDir + "/" + paramsFileName
	)

	return []byte(`local params = import "` + relComponentParamsPath + `";
params + {
  components +: {
    // Insert component parameter overrides here. Ex:
    // guestbook +: {
    //   name: "guestbook-dev",
    //   replicas: params.global.replicas,
    // },
  },
}
`)
}
