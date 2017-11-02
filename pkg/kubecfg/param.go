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
	"strconv"

	log "github.com/sirupsen/logrus"
)

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
