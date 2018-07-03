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
	"strings"

	"github.com/pkg/errors"

	"github.com/ksonnet/ksonnet/pkg/app"
	"github.com/ksonnet/ksonnet/pkg/component"
	"github.com/ksonnet/ksonnet/pkg/params"
	"github.com/ksonnet/ksonnet/pkg/pipeline"
	"github.com/ksonnet/ksonnet/pkg/util/table"
)

type findModuleFn func(a app.App, moduleName string) (component.Module, error)

// RunParamList runs `param list`.
func RunParamList(m map[string]interface{}) error {
	pl, err := NewParamList(m)
	if err != nil {
		return err
	}

	return pl.Run()
}

// paramsLister list params
type paramsLister interface {
	// List lists params given a source and optional component name.
	List(r io.Reader, componentName string) ([]params.Entry, error)
}

// ParamList lists parameters for a component.
type ParamList struct {
	app           app.App
	moduleName    string
	componentName string
	envName       string

	out          io.Writer
	findModuleFn findModuleFn

	modulesFn       func() ([]component.Module, error)
	envParametersFn func(string) (string, error)
	lister          paramsLister
}

// NewParamList creates an instances of ParamList.
func NewParamList(m map[string]interface{}) (*ParamList, error) {
	ol := newOptionLoader(m)

	pl := &ParamList{
		app:           ol.LoadApp(),
		moduleName:    ol.LoadOptionalString(OptionModule),
		componentName: ol.LoadOptionalString(OptionComponentName),
		envName:       ol.LoadOptionalString(OptionEnvName),

		out:          os.Stdout,
		findModuleFn: component.GetModule,
	}

	if ol.err != nil {
		return nil, ol.err
	}

	p := pipeline.New(pl.app, pl.envName)
	pl.modulesFn = p.Modules
	pl.envParametersFn = p.EnvParameters

	dest := app.EnvironmentDestinationSpec{}
	pl.lister = params.NewLister(pl.app.Root(), dest)

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

	r, err := module.ParamsSource()
	if err != nil {
		return errors.Wrap(err, "reading module parameters")
	}
	defer r.Close()

	entries, err := pl.lister.List(r, pl.componentName)
	if err != nil {
		return err
	}

	return pl.print(entries)
}

func (pl *ParamList) print(entries []params.Entry) error {
	table := table.New(pl.out)

	table.SetHeader([]string{"COMPONENT", "PARAM", "VALUE"})
	for _, entry := range entries {
		table.Append([]string{entry.ComponentName, entry.ParamName, entry.Value})
	}

	return table.Render()
}

func (pl *ParamList) handleEnvParams() error {
	modules, err := pl.modulesFn()
	if err != nil {
		return err
	}

	var entries []params.Entry
	for _, m := range modules {
		source, err := pl.envParametersFn(m.Name())
		if err != nil {
			return err
		}

		r := strings.NewReader(source)

		moduleEntries, err := pl.lister.List(r, pl.componentName)
		if err != nil {
			return err
		}

		entries = append(entries, moduleEntries...)
	}

	return pl.print(entries)
}
