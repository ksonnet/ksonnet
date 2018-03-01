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

package kubecfg

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/ksonnet/ksonnet/component"
	param "github.com/ksonnet/ksonnet/metadata/params"
	str "github.com/ksonnet/ksonnet/strings"
	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
	"github.com/spf13/afero"

	log "github.com/sirupsen/logrus"
)

func sortedKeys(params map[string]param.Params) []string {
	// keys maintains an alphabetically sorted list of the components
	keys := make([]string, 0, len(params))
	for key := range params {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func sortedParams(params param.Params) []string {
	// keys maintains an alphabetically sorted list of the params
	keys := make([]string, 0, len(params))
	for key := range params {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

// ----------------------------------------------------------------------------

// ParamSetCmd stores the information necessary to set component and
// environment params.
type ParamSetCmd struct {
	component string
	env       string

	param string
	value string
}

// NewParamSetCmd acts as a constructor for ParamSetCmd. It will also sanitize
// or "quote" the param value first if necessary.
func NewParamSetCmd(component, env, param, value string) *ParamSetCmd {
	return &ParamSetCmd{component: component, env: env, param: param, value: sanitizeParamValue(value)}
}

// Run executes the setting of params.
func (c *ParamSetCmd) Run() error {
	manager, err := manager()
	if err != nil {
		return err
	}

	if len(c.env) == 0 {
		if err = manager.SetComponentParams(c.component, param.Params{c.param: c.value}); err == nil {
			log.Infof("Parameter '%s' successfully set to '%s' for component '%s'", c.param, c.value, c.component)
		}
	} else {
		if err = manager.SetEnvironmentParams(c.env, c.component, param.Params{c.param: c.value}); err == nil {
			log.Infof("Parameter '%s' successfully set to '%s' for component '%s' in environment '%s'",
				c.param, c.value, c.component, c.env)
		}
	}

	return err
}

// sanitizeParamValue does a best effort to identify value types. It will put
// quotes around values which it categorizes as strings.
func sanitizeParamValue(value string) string {
	// numeric
	if _, err := strconv.ParseFloat(value, 64); err == nil {
		return value
	}
	// string
	return fmt.Sprintf(`"%s"`, value)
}

// ----------------------------------------------------------------------------

const (
	paramComponentHeader = "COMPONENT"
	paramNameHeader      = "PARAM"
	paramValueHeader     = "VALUE"
)

// ParamListCmd stores the information necessary display component or
// environment parameters
type ParamListCmd struct {
	fs        afero.Fs
	root      string
	component string
	env       string
	nsName    string
}

// NewParamListCmd acts as a constructor for ParamListCmd.
func NewParamListCmd(component, env, nsName string) *ParamListCmd {
	return &ParamListCmd{
		component: component,
		env:       env,
		nsName:    nsName,
	}
}

// Run executes the displaying of params.
func (c *ParamListCmd) Run(out io.Writer) error {
	cwd, err := os.Getwd()
	if err != nil {
		return errors.Wrap(err, "get current working directory")
	}

	manager, err := manager()
	if err != nil {
		return err
	}

	var params map[string]param.Params
	if len(c.env) != 0 {
		params, err = manager.GetEnvironmentParams(c.env, c.nsName)
		if err != nil {
			return err
		}
	} else {
		params, err = manager.GetAllComponentParams(cwd)
		if err != nil {
			return err
		}
	}

	if len(c.component) != 0 {
		if _, ok := params[c.component]; !ok {
			return fmt.Errorf("No such component '%s' found", c.component)
		}

		p := params[c.component]
		return outputParamsFor(c.component, p, out)
	}

	return outputParams(params, out)
}

func outputParamsFor(component string, params param.Params, out io.Writer) error {
	keys := sortedParams(params)

	rows := [][]string{
		[]string{paramNameHeader, paramValueHeader},
		[]string{strings.Repeat("=", len(paramNameHeader)), strings.Repeat("=", len(paramValueHeader))},
	}
	for _, k := range keys {
		rows = append(rows, []string{k, params[k]})
	}

	formatted, err := str.PadRows(rows)
	if err != nil {
		return err
	}
	_, err = fmt.Fprint(out, formatted)
	return err
}

func outputParams(params map[string]param.Params, out io.Writer) error {
	keys := sortedKeys(params)

	rows := [][]string{
		[]string{paramComponentHeader, paramNameHeader, paramValueHeader},
		[]string{
			strings.Repeat("=", len(paramComponentHeader)),
			strings.Repeat("=", len(paramNameHeader)),
			strings.Repeat("=", len(paramValueHeader))},
	}
	for _, k := range keys {
		// sort params to display alphabetically
		ps := sortedParams(params[k])

		for _, p := range ps {
			rows = append(rows, []string{k, p, params[k][p]})
		}
	}

	formatted, err := str.PadRows(rows)
	if err != nil {
		return err
	}
	_, err = fmt.Fprint(out, formatted)
	return err
}

// ----------------------------------------------------------------------------

// ParamDiffCmd stores the information necessary to diff between environment
// parameters.
type ParamDiffCmd struct {
	fs   afero.Fs
	root string
	env1 string
	env2 string

	component string
}

// NewParamDiffCmd acts as a constructor for ParamDiffCmd.
func NewParamDiffCmd(fs afero.Fs, root, env1, env2, componentName string) *ParamDiffCmd {
	return &ParamDiffCmd{
		fs:        fs,
		root:      root,
		env1:      env1,
		env2:      env2,
		component: componentName,
	}
}

type paramDiffRecord struct {
	component string
	param     string
	value1    string
	value2    string
}

// Run executes the diffing of environment params.
func (c *ParamDiffCmd) Run(out io.Writer) error {
	manager, err := manager()
	if err != nil {
		return err
	}

	ns, componentName := component.ExtractNamespacedComponent(c.fs, c.root, c.component)

	params1, err := manager.GetEnvironmentParams(c.env1, ns.Path)
	if err != nil {
		return err
	}

	params2, err := manager.GetEnvironmentParams(c.env2, ns.Path)
	if err != nil {
		return err
	}

	if len(c.component) != 0 {
		params1 = map[string]param.Params{componentName: params1[componentName]}
		params2 = map[string]param.Params{componentName: params2[componentName]}
	}

	if reflect.DeepEqual(params1, params2) {
		log.Info("No differences found.")
		return nil
	}

	componentNames := collectComponents(params1, params2)

	var rows [][]string
	for _, componentName := range componentNames {
		paramNames := collectParams(params1[componentName], params2[componentName])

		for _, paramName := range paramNames {
			var v1, v2 string
			var ok bool
			var p param.Params

			if p, ok = params1[componentName]; ok {
				v1 = p[paramName]
			}

			if p, ok = params2[componentName]; ok {
				v2 = p[paramName]
			}

			row := []string{
				componentName,
				paramName,
				v1,
				v2,
			}

			rows = append(rows, row)
		}
	}

	printTable([]string{"COMPONENT", "PARAM", c.env1, c.env2}, rows)
	return nil
}

func collectComponents(param1, param2 map[string]param.Params) []string {
	m := make(map[string]bool)
	for k := range param1 {
		m[k] = true
	}
	for k := range param2 {
		m[k] = true
	}

	var names []string

	for k := range m {
		names = append(names, k)
	}

	sort.Strings(names)

	return names
}

func collectParams(param1, param2 param.Params) []string {
	m := make(map[string]bool)
	for k := range param1 {
		m[k] = true
	}
	for k := range param2 {
		m[k] = true
	}

	var names []string

	for k := range m {
		names = append(names, k)
	}

	sort.Strings(names)

	return names
}

func printTable(headers []string, data [][]string) {
	headerLens := make([]int, len(headers))
	for i := range headers {
		headerLens[i] = len(headers[i])
	}

	for i := range headerLens {
		headers[i] = fmt.Sprintf("%s\n%s", headers[i], strings.Repeat("=", headerLens[i]))
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(headers)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetRowLine(false)
	table.SetBorder(false)
	table.AppendBulk(data)
	table.Render()
}
