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

package actions

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/ksonnet/ksonnet/pkg/app"
	amocks "github.com/ksonnet/ksonnet/pkg/app/mocks"
	"github.com/ksonnet/ksonnet/pkg/component"
	cmocks "github.com/ksonnet/ksonnet/pkg/component/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParamList(t *testing.T) {
	moduleParams := []component.ModuleParameter{
		{Component: "deployment", Index: "0", Key: "key", Value: `"value"`},
	}

	module := &cmocks.Module{}
	module.On("Params", "envName").Return(moduleParams, nil)
	module.On("Params", "").Return(moduleParams, nil)

	c := &cmocks.Component{}
	c.On("Params", "").Return(moduleParams, nil)

	withApp(t, func(appMock *amocks.App) {
		cases := []struct {
			name            string
			in              map[string]interface{}
			findModulesFn   func(t *testing.T) findModulesFn
			findModuleFn    func(t *testing.T) findModuleFn
			findComponentFn func(t *testing.T) findComponentFn
			outputFile      string
		}{
			{
				name: "component name",
				in: map[string]interface{}{
					OptionApp:           appMock,
					OptionComponentName: "deployment",
					OptionModule:        "module",
				},
				findModuleFn: func(t *testing.T) findModuleFn {
					return func(a app.App, moduleName string) (component.Module, error) {
						assert.Equal(t, "module", moduleName)
						return module, nil
					}
				},
				findComponentFn: func(t *testing.T) findComponentFn {
					return func(a app.App, moduleName, componentName string) (component.Component, error) {
						assert.Equal(t, "module", moduleName)
						assert.Equal(t, "deployment", componentName)
						return c, nil
					}
				},
				outputFile: filepath.Join("param", "list", "with_component.txt"),
			},
			{
				name: "no component name",
				in: map[string]interface{}{
					OptionApp:    appMock,
					OptionModule: "module",
				},
				findModuleFn: func(t *testing.T) findModuleFn {
					return func(a app.App, moduleName string) (component.Module, error) {
						assert.Equal(t, "module", moduleName)
						return module, nil
					}
				},
				findComponentFn: func(t *testing.T) findComponentFn {
					return func(a app.App, moduleName, componentName string) (component.Component, error) {
						assert.Equal(t, "module", moduleName)
						assert.Equal(t, "deployment", componentName)
						return c, nil
					}
				},
				outputFile: filepath.Join("param", "list", "without_component.txt"),
			},
			{
				name: "env",
				in: map[string]interface{}{
					OptionApp:     appMock,
					OptionEnvName: "envName",
				},
				findModulesFn: func(t *testing.T) findModulesFn {
					return func(a app.App, envName string) ([]component.Module, error) {
						assert.Equal(t, "envName", envName)
						return []component.Module{module}, nil
					}
				},
				outputFile: filepath.Join("param", "list", "env.txt"),
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {

				a, err := NewParamList(tc.in)
				require.NoError(t, err)

				if tc.findModulesFn != nil {
					a.findModulesFn = tc.findModulesFn(t)
				}

				if tc.findModuleFn != nil {
					a.findModuleFn = tc.findModuleFn(t)
				}

				if tc.findComponentFn != nil {
					a.findComponentFn = tc.findComponentFn(t)
				}

				var buf bytes.Buffer
				a.out = &buf

				err = a.Run()
				require.NoError(t, err)

				assertOutput(t, tc.outputFile, buf.String())
			})
		}
	})
}
