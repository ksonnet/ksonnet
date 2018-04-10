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
	"io"
	"os"

	"github.com/ksonnet/ksonnet/component"
	"github.com/ksonnet/ksonnet/metadata/app"
	"github.com/ksonnet/ksonnet/pkg/util/table"
	"github.com/pkg/errors"
)

type findModulesFn func(a app.App, envName string) ([]component.Module, error)
type findModuleFn func(a app.App, moduleName string) (component.Module, error)
type findComponentFn func(a app.App, moduleName, componentName string) (component.Component, error)

// RunParamList runs `param list`.
func RunParamList(m map[string]interface{}) error {
	pl, err := NewParamList(m)
	if err != nil {
		return err
	}

	return pl.Run()
}

// ParamList lists parameters for a component.
type ParamList struct {
	app           app.App
	moduleName    string
	componentName string
	envName       string

	out             io.Writer
	findModulesFn   findModulesFn
	findModuleFn    findModuleFn
	findComponentFn findComponentFn
}

// NewParamList creates an instances of ParamList.
func NewParamList(m map[string]interface{}) (*ParamList, error) {
	ol := newOptionLoader(m)

	pl := &ParamList{
		app:           ol.loadApp(),
		moduleName:    ol.loadOptionalString(OptionModule),
		componentName: ol.loadOptionalString(OptionComponentName),
		envName:       ol.loadOptionalString(OptionEnvName),

		out:             os.Stdout,
		findModulesFn:   component.ModulesFromEnv,
		findModuleFn:    component.GetModule,
		findComponentFn: component.LocateComponent,
	}

	if ol.err != nil {
		return nil, ol.err
	}

	return pl, nil
}

// Run runs the ParamList action.
func (pl *ParamList) Run() error {
	if pl.envName != "" {
		return pl.handleEnvParams()
	}

	module, err := pl.findModuleFn(pl.app, pl.moduleName)
	if err != nil {
		return errors.Wrap(err, "could not find module")
	}

	params, err := pl.collectParams(module)
	if err != nil {
		return err
	}

	return pl.print(params)
}

func (pl *ParamList) handleEnvParams() error {
	modules, err := pl.findModulesFn(pl.app, pl.envName)
	if err != nil {
		return err
	}

	var params []component.ModuleParameter

	for _, module := range modules {
		moduleParams, err := pl.collectParams(module)
		if err != nil {
			return err
		}

		if pl.moduleName != "" && module.Name() != pl.moduleName {
			continue
		}
		params = append(params, moduleParams...)
	}

	return pl.print(params)

}

func (pl *ParamList) collectParams(module component.Module) ([]component.ModuleParameter, error) {
	if pl.componentName == "" {
		return module.Params(pl.envName)
	}

	c, err := pl.findComponentFn(pl.app, pl.moduleName, pl.componentName)
	if err != nil {
		return nil, err
	}

	return c.Params(pl.envName)
}

func (pl *ParamList) print(params []component.ModuleParameter) error {
	table := table.New(pl.out)

	table.SetHeader([]string{"COMPONENT", "INDEX", "PARAM", "VALUE"})
	for _, data := range params {
		table.Append([]string{data.Component, data.Index, data.Key, data.Value})
	}

	return table.Render()
}
