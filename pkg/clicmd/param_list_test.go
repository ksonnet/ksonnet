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

package clicmd

import (
	"testing"

	"github.com/ksonnet/ksonnet/pkg/actions"
)

func Test_paramListCmd(t *testing.T) {
	cases := []cmdTestCase{
		{
			name:   "in general",
			args:   []string{"param", "list"},
			action: actionParamList,
			expected: map[string]interface{}{
				actions.OptionApp:            nil,
				actions.OptionEnvName:        "",
				actions.OptionModule:         "",
				actions.OptionComponentName:  "",
				actions.OptionOutput:         "",
				actions.OptionWithoutModules: false,
			},
		},
		{
			name:   "with output",
			args:   []string{"param", "list", "-o", "json"},
			action: actionParamList,
			expected: map[string]interface{}{
				actions.OptionApp:            nil,
				actions.OptionEnvName:        "",
				actions.OptionModule:         "",
				actions.OptionComponentName:  "",
				actions.OptionOutput:         "json",
				actions.OptionWithoutModules: false,
			},
		},
		{
			name:   "component",
			args:   []string{"param", "list", "component"},
			action: actionParamList,
			expected: map[string]interface{}{
				actions.OptionApp:            nil,
				actions.OptionEnvName:        "",
				actions.OptionModule:         "",
				actions.OptionComponentName:  "component",
				actions.OptionOutput:         "",
				actions.OptionWithoutModules: false,
			},
		},
		{
			name:   "component",
			args:   []string{"param", "list", "--module", "module"},
			action: actionParamList,
			expected: map[string]interface{}{
				actions.OptionApp:            nil,
				actions.OptionEnvName:        "",
				actions.OptionModule:         "module",
				actions.OptionComponentName:  "",
				actions.OptionOutput:         "",
				actions.OptionWithoutModules: false,
			},
		},
		{
			name:   "env",
			args:   []string{"param", "list", "--env", "env"},
			action: actionParamList,
			expected: map[string]interface{}{
				actions.OptionApp:            nil,
				actions.OptionEnvName:        "env",
				actions.OptionModule:         "",
				actions.OptionComponentName:  "",
				actions.OptionOutput:         "",
				actions.OptionWithoutModules: false,
			},
		},
		{
			name:   "env without modules",
			args:   []string{"param", "list", "--env", "env", "--without-modules"},
			action: actionParamList,
			expected: map[string]interface{}{
				actions.OptionApp:            nil,
				actions.OptionEnvName:        "env",
				actions.OptionModule:         "",
				actions.OptionComponentName:  "",
				actions.OptionOutput:         "",
				actions.OptionWithoutModules: true,
			},
		},
	}

	runTestCmd(t, cases)
}
