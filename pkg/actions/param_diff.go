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

	"github.com/ksonnet/ksonnet/pkg/app"
	"github.com/ksonnet/ksonnet/pkg/component"
	"github.com/ksonnet/ksonnet/pkg/util/table"
)

// RunParamDiff runs `param diff`.
func RunParamDiff(m map[string]interface{}) error {
	pd, err := NewParamDiff(m)
	if err != nil {
		return err
	}

	return pd.Run()
}

// ParamDiff shows difference between params in two environments.
type ParamDiff struct {
	app           app.App
	envName1      string
	envName2      string
	componentName string

	modulesFromEnvFn func(app.App, string) ([]component.Module, error)
	out              io.Writer
}

// NewParamDiff creates an instance of ParamDiff.
func NewParamDiff(m map[string]interface{}) (*ParamDiff, error) {
	ol := newOptionLoader(m)

	pd := &ParamDiff{
		app:           ol.LoadApp(),
		envName1:      ol.LoadString(OptionEnvName1),
		envName2:      ol.LoadString(OptionEnvName2),
		componentName: ol.LoadOptionalString(OptionComponentName),

		modulesFromEnvFn: component.ModulesFromEnv,
		out:              os.Stdout,
	}

	if ol.err != nil {
		return nil, ol.err
	}

	return pd, nil
}

// Run runs the action.
func (pd *ParamDiff) Run() error {
	env1Params, err := pd.moduleParams(pd.envName1)
	if err != nil {
		return err
	}

	env2Params, err := pd.moduleParams(pd.envName2)
	if err != nil {
		return err
	}

	var rows [][]string

	rows = append(rows, pd.checkDiff(env1Params, env2Params)...)
	rows = append(rows, pd.checkMissing(env1Params, env2Params)...)

	return pd.print(rows)
}

func (pd *ParamDiff) checkDiff(env1, env2 []component.ModuleParameter) [][]string {
	var rows [][]string

	for _, mp1 := range env1 {
		if pd.componentName != "" && pd.componentName != mp1.Component {
			continue
		}
		for _, mp2 := range env2 {
			if mp1.IsSameType(mp2) && mp1.Value != mp2.Value {
				rows = append(rows, []string{mp1.Component, mp1.Index, mp1.Key, mp1.Value, mp2.Value})
			}
		}
	}

	return rows
}

// nolint: gocyclo
func (pd *ParamDiff) checkMissing(env1, env2 []component.ModuleParameter) [][]string {
	var rows [][]string

	for _, mp1 := range env1 {
		if pd.componentName != "" && pd.componentName != mp1.Component {
			continue
		}
		found := false
		for _, mp2 := range env2 {
			if mp1.IsSameType(mp2) {
				found = true
			}
		}

		if !found {
			rows = append(rows, []string{mp1.Component, mp1.Index, mp1.Key, mp1.Value})
		}
	}

	for _, mp1 := range env2 {
		if pd.componentName != "" && pd.componentName != mp1.Component {
			continue
		}
		found := false
		for _, mp2 := range env1 {
			if mp1.IsSameType(mp2) {
				found = true
			}
		}

		if !found {
			rows = append(rows, []string{mp1.Component, mp1.Index, mp1.Key, "", mp1.Value})
		}
	}

	return rows
}

func (pd *ParamDiff) moduleParams(envName string) ([]component.ModuleParameter, error) {
	modules, err := pd.modulesFromEnvFn(pd.app, envName)
	if err != nil {
		return nil, err
	}

	var moduleParams []component.ModuleParameter
	for _, module := range modules {
		p, err := module.Params(envName)
		if err != nil {
			return nil, err
		}

		moduleParams = append(moduleParams, p...)
	}

	return moduleParams, nil
}

func (pd *ParamDiff) print(rows [][]string) error {
	table := table.New(pd.out)

	table.SetHeader([]string{"COMPONENT", "INDEX", "PARAM", "ENV1", "ENV2"})
	table.AppendBulk(rows)

	return table.Render()
}
