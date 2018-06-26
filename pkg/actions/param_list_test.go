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
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ksonnet/ksonnet/pkg/app"
	amocks "github.com/ksonnet/ksonnet/pkg/app/mocks"
	"github.com/ksonnet/ksonnet/pkg/component"
	cmocks "github.com/ksonnet/ksonnet/pkg/component/mocks"
	"github.com/ksonnet/ksonnet/pkg/params"
	paramsTesting "github.com/ksonnet/ksonnet/pkg/params/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParamList(t *testing.T) {
	moduleParams := []component.ModuleParameter{
		{Component: "deployment", Key: "key", Value: `"value"`},
	}

	module := &cmocks.Module{}

	p := `{components:{deployment:{key:"value"}}}`
	paramsFile := ioutil.NopCloser(strings.NewReader(p))
	module.On("ParamsSource").Return(paramsFile, nil)

	c := &cmocks.Component{}
	c.On("Params", "").Return(moduleParams, nil)

	fakeLister := &paramsTesting.FakeLister{
		Entries: []params.Entry{
			{ComponentName: "deployment", ParamName: "key", Value: `'value'`},
		},
	}

	withApp(t, func(appMock *amocks.App) {
		ec := &app.EnvironmentConfig{}
		appMock.On("Environment", "envName").Return(ec, nil)

		cases := []struct {
			name         string
			in           map[string]interface{}
			findModuleFn func(t *testing.T) findModuleFn
			modulesFn    func() ([]component.Module, error)
			lister       paramsLister
			outputFile   string
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
				lister:     fakeLister,
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
				lister:     fakeLister,
				outputFile: filepath.Join("param", "list", "without_component.txt"),
			},
			{
				name: "env",
				in: map[string]interface{}{
					OptionApp:     appMock,
					OptionEnvName: "envName",
				},
				modulesFn: func() ([]component.Module, error) {
					module.On("Name").Return("/")
					return []component.Module{module}, nil
				},
				lister:     fakeLister,
				outputFile: filepath.Join("param", "list", "env.txt"),
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				a, err := NewParamList(tc.in)
				require.NoError(t, err)

				a.lister = tc.lister

				if tc.findModuleFn != nil {
					a.findModuleFn = tc.findModuleFn(t)
				}

				if tc.modulesFn != nil {
					a.modulesFn = tc.modulesFn
				}

				a.envParametersFn = func(string) (string, error) {
					return "{}", nil
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

func TestParamList_requires_app(t *testing.T) {
	in := make(map[string]interface{})
	_, err := NewParamList(in)
	require.Error(t, err)
}
