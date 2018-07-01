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

package params

import (
	"path/filepath"

	"github.com/ksonnet/ksonnet/pkg/app"
	"github.com/ksonnet/ksonnet/pkg/util/jsonnet"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
)

// EvaluateEnv evaluates environment parameters.
func EvaluateEnv(a app.App, sourcePath, paramsStr, envName, moduleName string) (string, error) {
	snippet, err := afero.ReadFile(a.Fs(), sourcePath)
	if err != nil {
		return "", err
	}

	paramsStr, err = modularizeParameters(a, envName, moduleName, paramsStr)
	if err != nil {
		return "", errors.Wrap(err, "modularizing parameters")
	}

	moduleParams, err := BuildEnvParamsForModule(moduleName, string(snippet), paramsStr)
	if err != nil {
		return "", errors.Wrapf(err, "selecting params for module %q in environment %q", moduleName, envName)
	}

	envParams, err := evaluateEnvInVM(a, envName, sourcePath, moduleParams, paramsStr)
	if err != nil {
		return "", errors.Wrapf(err, "evaluating parameters for module %q in environment %q", moduleName, envName)
	}

	return envParams, nil
}

// modularizeParameters adds a module prefix to component parameters.
// * Given a root module, it will not update the component name
// * Given a module nested under root, it will prepend the module: eg: `module apps -> apps.component`
func modularizeParameters(a app.App, envName, moduleName, paramsStr string) (string, error) {
	script, err := loadScript("modularize_params.libsonnet")
	if err != nil {
		return "", errors.Wrap(err, "loading script")
	}

	envCode, err := JsonnetEnvObject(a, envName)
	if err != nil {
		return "", errors.Wrap(err, "generating environments object")
	}

	vm := jsonnet.NewVM()
	vm.ExtCode("__ksonnet/environments", envCode)
	vm.TLAVar("moduleName", moduleName)
	vm.TLACode("params", paramsStr)

	output, err := vm.EvaluateSnippet("modularize-params", script)
	if err != nil {
		return "", errors.Wrap(err, "adding module to params")
	}

	return output, nil
}

// evaluateEnvInVM evaluates the environment parameters in a Jsonnet VM.
func evaluateEnvInVM(a app.App, envName, sourcePath, snippet, paramsStr string) (string, error) {
	libPath, err := a.LibPath(envName)
	if err != nil {
		return "", err
	}

	vm := jsonnet.NewVM()

	vm.AddJPath(
		libPath,
		filepath.Join(a.Root(), "lib"),
		filepath.Join(a.Root(), "vendor"),
	)
	vm.ExtCode("__ksonnet/params", paramsStr)

	return vm.EvaluateSnippet(sourcePath, snippet)
}
