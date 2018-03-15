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

package component

import (
	"path/filepath"
	"sort"

	"github.com/ksonnet/ksonnet/metadata/app"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
)

type componentPathLocator struct {
	app     app.App
	envSpec *app.EnvironmentSpec
}

func newComponentPathLocator(a app.App, envName string) (*componentPathLocator, error) {
	if a == nil {
		return nil, errors.New("app is nil")
	}

	env, err := a.Environment(envName)
	if err != nil {
		return nil, err
	}

	return &componentPathLocator{
		app:     a,
		envSpec: env,
	}, nil
}

func (cpl *componentPathLocator) Locate() ([]string, error) {
	targets := cpl.envSpec.Targets
	rootPath := cpl.app.Root()

	if len(targets) == 0 {
		return []string{filepath.Join(rootPath, componentsRoot)}, nil
	}

	var paths []string

	for _, target := range targets {
		childPath := filepath.Join(rootPath, componentsRoot, target)
		exists, err := afero.DirExists(cpl.app.Fs(), childPath)
		if err != nil {
			return nil, err
		}

		if !exists {
			return nil, errors.Errorf("target %q is not valid", target)
		}

		paths = append(paths, childPath)
	}

	sort.Strings(paths)

	return paths, nil
}

// isComponent reports if a file is a component. Components have a `jsonnet` extension.
func isComponent(path string) bool {
	for _, s := range []string{".jsonnet", ".yaml", "json"} {
		if s == filepath.Ext(path) {
			return true
		}
	}
	return false
}
