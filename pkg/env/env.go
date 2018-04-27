// Copyright 2018 The kubecfg authors
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

package env

import (
	"encoding/json"
	"path/filepath"

	"github.com/ksonnet/ksonnet/pkg/app"
	"github.com/ksonnet/ksonnet/pkg/util/jsonnet"
	"github.com/spf13/afero"
)

const (
	// primary environment files.
	envFileName     = "main.jsonnet"
	paramsFileName  = "params.libsonnet"
	globalsFileName = "globals.libsonnet"

	// envRootName is the name for the environment root.
	envRootName = "environments"
)

var (
	componentJPaths  = make([]string, 0)
	componentExtVars = make(map[string]string)
	componentTlaVars = make(map[string]string)
)

// AddJPaths adds paths to JPath for a component evaluation.
func AddJPaths(paths ...string) {
	componentJPaths = append(componentJPaths, paths...)
}

// AddExtVar adds an ext var to a component evaluation.
func AddExtVar(key, value string) {
	componentExtVars[key] = value
}

// AddExtVarFile adds an ext var from a file to component evaluation.
func AddExtVarFile(a app.App, key, filePath string) error {
	data, err := afero.ReadFile(a.Fs(), filePath)
	if err != nil {
		return err
	}

	componentExtVars[key] = string(data)
	return nil
}

// AddTlaVar adds a tla var to a component evaluation.
func AddTlaVar(key, value string) {
	componentTlaVars[key] = value
}

// AddTlaVarFile adds a tla var from a file to component evaluation.
func AddTlaVarFile(a app.App, key, filePath string) error {
	data, err := afero.ReadFile(a.Fs(), filePath)
	if err != nil {
		return err
	}

	componentTlaVars[key] = string(data)
	return nil
}

// MainFile returns the contents of the environment's main source.
func MainFile(a app.App, envName string) (string, error) {
	path, err := Path(a, envName, envFileName)
	if err != nil {
		return "", err
	}

	snippet, err := afero.ReadFile(a.Fs(), path)
	if err != nil {
		return "", err
	}

	return string(snippet), nil
}

// Evaluate evaluates an environment.
func Evaluate(a app.App, envName, components, paramsStr string) (string, error) {
	libPath, err := a.LibPath(envName)
	if err != nil {
		return "", err
	}

	appEnv, err := a.Environment(envName)
	if err != nil {
		return "", err
	}

	snippet, err := MainFile(a, envName)
	if err != nil {
		return "", err
	}

	vm := jsonnet.NewVM()
	vm.AddJPath(componentJPaths...)
	vm.AddJPath(
		filepath.Join(a.Root(), envRootName),
		filepath.Join(a.Root(), envRootName, appEnv.Path),
		filepath.Join(a.Root(), "vendor"),
		libPath,
	)

	envCode, err := environmentsCode(a, envName)
	if err != nil {
		return "", err
	}

	for k, v := range componentExtVars {
		vm.ExtVar(k, v)
	}

	for k, v := range componentTlaVars {
		vm.TLAVar(k, v)
	}

	vm.ExtCode("__ksonnet/environments", envCode)
	vm.ExtCode(ComponentsExtCodeKey, components)
	vm.ExtCode("__ksonnet/params", paramsStr)

	return vm.EvaluateSnippet(envFileName, snippet)
}

func envRoot(a app.App, envName string) (string, error) {
	envSpec, err := a.Environment(envName)
	if err != nil {
		return "", err
	}

	return filepath.Join(a.Root(), envRootName, envSpec.Path), nil

}

func Path(a app.App, envName string, path ...string) (string, error) {
	base, err := envRoot(a, envName)
	if err != nil {
		return "", err
	}

	return filepath.Join(append([]string{base}, path...)...), nil
}

func environmentsCode(a app.App, envName string) (string, error) {
	envDetails, err := a.Environment(envName)
	if err != nil {
		return "", err
	}

	dest := map[string]string{
		"server":    envDetails.Destination.Server,
		"namespace": envDetails.Destination.Namespace,
	}

	marshalledDestination, err := json.Marshal(&dest)
	if err != nil {
		return "", err
	}

	return string(marshalledDestination), nil
}
