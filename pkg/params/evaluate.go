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
	"github.com/spf13/afero"
)

// EvaluateEnv evaluates an env with jsonnet.
func EvaluateEnv(a app.App, sourcePath, paramsStr, envName string) (string, error) {
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

	snippet, err := afero.ReadFile(a.Fs(), sourcePath)
	if err != nil {
		return "", err
	}

	return vm.EvaluateSnippet(sourcePath, string(snippet))
}
