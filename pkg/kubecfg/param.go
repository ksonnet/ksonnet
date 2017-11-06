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
	"reflect"
	"sort"
	"strconv"
	"strings"

	param "github.com/ksonnet/ksonnet/metadata/params"

	"github.com/fatih/color"
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
	// boolean
	if value == "true" || value == "false" {
		return value
	}
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
	component string
	env       string
}

// NewParamListCmd acts as a constructor for ParamListCmd.
func NewParamListCmd(component, env string) *ParamListCmd {
	return &ParamListCmd{component: component, env: env}
}

// Run executes the displaying of params.
func (c *ParamListCmd) Run(out io.Writer) error {
	manager, err := manager()
	if err != nil {
		return err
	}

	var params map[string]param.Params
	if len(c.env) != 0 {
		params, err = manager.GetEnvironmentParams(c.env)
		if err != nil {
			return err
		}
	} else {
		params, err = manager.GetAllComponentParams()
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

	//
	// Format each parameter information for pretty printing.
	// Each parameter will be outputted alphabetically like the following:
	//
	//   PARAM    VALUE
	//   name     "foo"
	//   replicas 1
	//

	maxParamLen := len(paramNameHeader)
	for _, k := range keys {
		if l := len(k); l > maxParamLen {
			maxParamLen = l
		}
	}

	nameSpacing := strings.Repeat(" ", maxParamLen-len(paramNameHeader)+1)

	lines := []string{}
	lines = append(lines, paramNameHeader+nameSpacing+paramValueHeader+"\n")
	lines = append(lines, strings.Repeat("=", len(paramNameHeader))+nameSpacing+
		strings.Repeat("=", len(paramValueHeader))+"\n")

	for _, k := range keys {
		nameSpacing = strings.Repeat(" ", maxParamLen-len(k)+1)
		lines = append(lines, k+nameSpacing+params[k]+"\n")
	}

	_, err := fmt.Fprint(out, strings.Join(lines, ""))
	return err
}

func outputParams(params map[string]param.Params, out io.Writer) error {
	keys := sortedKeys(params)

	//
	// Format each component parameter information for pretty printing.
	// Each component will be outputted alphabetically like the following:
	//
	//   COMPONENT PARAM     VALUE
	//   bar       name      "bar"
	//   bar       replicas  2
	//   foo       name      "foo"
	//   foo       replicas  1
	//

	maxComponentLen := len(paramComponentHeader)
	for _, k := range keys {
		if l := len(k); l > maxComponentLen {
			maxComponentLen = l
		}
	}

	maxParamLen := len(paramNameHeader) + maxComponentLen + 1
	for _, k := range keys {
		for p := range params[k] {
			if l := len(p) + maxComponentLen + 1; l > maxParamLen {
				maxParamLen = l
			}
		}
	}

	componentSpacing := strings.Repeat(" ", maxComponentLen-len(paramComponentHeader)+1)
	nameSpacing := strings.Repeat(" ", maxParamLen-maxComponentLen-len(paramNameHeader))

	lines := []string{}
	lines = append(lines, paramComponentHeader+componentSpacing+paramNameHeader+nameSpacing+paramValueHeader+"\n")
	lines = append(lines, strings.Repeat("=", len(paramComponentHeader))+componentSpacing+
		strings.Repeat("=", len(paramNameHeader))+nameSpacing+strings.Repeat("=", len(paramValueHeader))+"\n")

	for _, k := range keys {
		// sort params to display alphabetically
		ps := sortedParams(params[k])

		for _, p := range ps {
			componentSpacing = strings.Repeat(" ", maxComponentLen-len(k)+1)
			nameSpacing = strings.Repeat(" ", maxParamLen-maxComponentLen-len(p))
			lines = append(lines, k+componentSpacing+p+nameSpacing+params[k][p]+"\n")
		}
	}

	_, err := fmt.Fprint(out, strings.Join(lines, ""))
	return err
}

// ----------------------------------------------------------------------------

// ParamDiffCmd stores the information necessary to diff between environment
// parameters.
type ParamDiffCmd struct {
	env1 string
	env2 string

	component string
}

// NewParamDiffCmd acts as a constructor for ParamDiffCmd.
func NewParamDiffCmd(env1, env2, component string) *ParamDiffCmd {
	return &ParamDiffCmd{env1: env1, env2: env2, component: component}
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

	params1, err := manager.GetEnvironmentParams(c.env1)
	if err != nil {
		return err
	}

	params2, err := manager.GetEnvironmentParams(c.env2)
	if err != nil {
		return err
	}

	if len(c.component) != 0 {
		params1 = map[string]param.Params{c.component: params1[c.component]}
		params2 = map[string]param.Params{c.component: params2[c.component]}
	}

	if reflect.DeepEqual(params1, params2) {
		log.Info("No differences found.")
		return nil
	}

	records := diffParams(params1, params2)

	//
	// Format each component parameter information for pretty printing.
	// Each component will be outputted alphabetically like the following:
	//
	//   COMPONENT PARAM     dev       prod
	//   bar       name      "bar-dev" "bar"
	//   foo       replicas  1
	//

	maxComponentLen := len(paramComponentHeader)
	for _, k := range records {
		if l := len(k.component); l > maxComponentLen {
			maxComponentLen = l
		}
	}

	maxParamLen := len(paramNameHeader) + maxComponentLen + 1
	for _, k := range records {
		if l := len(k.param) + maxComponentLen + 1; l > maxParamLen {
			maxParamLen = l
		}
	}

	maxEnvLen := len(c.env1) + maxParamLen + 1
	for _, k := range records {
		if l := len(k.value1) + maxParamLen + 1; l > maxEnvLen {
			maxEnvLen = l
		}
	}

	componentSpacing := strings.Repeat(" ", maxComponentLen-len(paramComponentHeader)+1)
	nameSpacing := strings.Repeat(" ", maxParamLen-maxComponentLen-len(paramNameHeader))
	envSpacing := strings.Repeat(" ", maxEnvLen-maxParamLen-len(c.env1))

	// print headers
	color.New(color.FgBlack).Fprintln(out, paramComponentHeader+componentSpacing+
		paramNameHeader+nameSpacing+c.env1+envSpacing+c.env2)
	color.New(color.FgBlack).Fprintln(out, strings.Repeat("=", len(paramComponentHeader))+componentSpacing+
		strings.Repeat("=", len(paramNameHeader))+nameSpacing+
		strings.Repeat("=", len(c.env1))+envSpacing+
		strings.Repeat("=", len(c.env2)))

	// print body
	for _, k := range records {
		componentSpacing = strings.Repeat(" ", maxComponentLen-len(k.component)+1)
		nameSpacing = strings.Repeat(" ", maxParamLen-maxComponentLen-len(k.param))
		envSpacing = strings.Repeat(" ", maxEnvLen-maxParamLen-len(k.value1))
		line := fmt.Sprint(k.component + componentSpacing + k.param + nameSpacing + k.value1 + envSpacing + k.value2)
		if len(k.value1) == 0 {
			color.New(color.FgGreen).Fprintln(out, line)
		} else if len(k.value2) == 0 {
			color.New(color.FgRed).Fprintln(out, line)
		} else if k.value1 != k.value2 {
			color.New(color.FgYellow).Fprintln(out, line)
		} else {
			color.New(color.FgBlack).Fprintln(out, line)
		}
	}

	return nil
}

func diffParams(params1, params2 map[string]param.Params) []*paramDiffRecord {
	var records []*paramDiffRecord

	for c := range params1 {
		if _, contains := params2[c]; !contains {
			// env2 doesn't have this component, add all params from env1 for this component
			for p := range params2[c] {
				records = addRecord(records, c, p, params1[c][p], "")
			}
		} else {
			// has same component -- need to compare params
			for p := range params1[c] {
				if _, hasParam := params2[c][p]; !hasParam {
					// env2 doesn't have this param, add a record with the param value from env1
					records = addRecord(records, c, p, params1[c][p], "")
				} else {
					// env2 has this param too, add a record with both param values
					records = addRecord(records, c, p, params1[c][p], params2[c][p])
				}
			}
			// add remaining records for params that env2 has that env1 does not for this component
			for p := range params2[c] {
				if _, hasParam := params1[c][p]; !hasParam {
					records = addRecord(records, c, p, "", params2[c][p])
				}
			}
		}
	}

	// add remaining records where env2 contains a component that env1 does not
	for c := range params2 {
		if _, contains := params1[c]; !contains {
			for p := range params2[c] {
				records = addRecord(records, c, p, "", params2[c][p])
			}
		}
	}

	sort.Slice(records, func(i, j int) bool {
		if records[i].component == records[j].component {
			return records[i].param < records[j].param
		}
		return records[i].component < records[j].component
	})

	return records
}

func addRecord(records []*paramDiffRecord, component, param, value1, value2 string) []*paramDiffRecord {
	records = append(records, &paramDiffRecord{
		component: component,
		param:     param,
		value1:    value1,
		value2:    value2,
	})
	return records
}
