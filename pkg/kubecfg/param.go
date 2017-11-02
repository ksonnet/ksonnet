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
	"sort"
	"strconv"
	"strings"

	param "github.com/ksonnet/ksonnet/metadata/params"

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
