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

	"github.com/fatih/color"
	"github.com/ksonnet/ksonnet/component"
	param "github.com/ksonnet/ksonnet/metadata/params"
	"github.com/ksonnet/ksonnet/pkg/util/table"
	str "github.com/ksonnet/ksonnet/strings"
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
		outputParamsFor(c.component, p, out)
		return nil
	}

	outputParams(params, out)
	return nil
}

func outputParamsFor(component string, params param.Params, out io.Writer) {
	keys := sortedParams(params)

	t := table.New(out)
	t.SetHeader([]string{paramNameHeader, paramValueHeader})
	for _, k := range keys {
		t.Append([]string{k, params[k]})
	}

	t.Render()
}

func outputParams(params map[string]param.Params, out io.Writer) {
	keys := sortedKeys(params)

	t := table.New(out)
	t.SetHeader([]string{paramComponentHeader, paramNameHeader, paramValueHeader})

	for _, k := range keys {
		// sort params to display alphabetically
		ps := sortedParams(params[k])

		for _, p := range ps {
			t.Append([]string{k, p, params[k][p]})
		}
	}

	t.Render()
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
	const (
		componentHeader = "COMPONENT"
		paramHeader     = "PARAM"
	)

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

	headers := str.Row{
		Content: []string{componentHeader, paramHeader, c.env1, c.env2},
	}

	var body []str.Row
	for _, componentName := range componentNames {
		paramNames := collectParams(params1[componentName], params2[componentName])

		for _, paramName := range paramNames {
			var v1, v2 string

			if p, ok := params1[componentName]; ok {
				v1 = p[paramName]
			}

			if p, ok := params2[componentName]; ok {
				v2 = p[paramName]
			}

			var bgColor *color.Color
			if v1 == "" {
				bgColor = color.New(color.BgGreen)
			} else if v2 == "" {
				bgColor = color.New(color.BgRed)
			} else if v1 != v2 {
				bgColor = color.New(color.BgYellow)
			}

			body = append(body, str.Row{
				Content: []string{
					componentName,
					paramName,
					v1,
					v2,
				},
				Color: bgColor,
			})
		}
	}

	formatted, err := str.Table(headers, body)
	if err != nil {
		return err
	}

	for _, row := range formatted {
		if row.Color != nil {
			_, err = row.Color.Fprint(out, row.Content)
			if err != nil {
				return err
			}
			// Must print new line separately otherwise color alignment will be
			// incorrect.
			fmt.Println()
		} else {
			_, err = fmt.Fprintln(out, row.Content)
			if err != nil {
				return err
			}
		}
	}

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
