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

func Test_registryUpdateCmd(t *testing.T) {
	cases := []cmdTestCase{
		{
			name:   "registry",
			args:   []string{"registry", "update", "databases"},
			action: actionRegistryUpdate,
			expected: map[string]interface{}{
				actions.OptionApp:     nil,
				actions.OptionName:    "databases",
				actions.OptionVersion: "",
			},
		},
		{
			name:   "registry with version",
			args:   []string{"registry", "update", "databases", "0.0.1"},
			action: actionRegistryUpdate,
			expected: map[string]interface{}{
				actions.OptionApp:     nil,
				actions.OptionName:    "databases",
				actions.OptionVersion: "0.0.1",
			},
		},
		{
			name:  "invalid arguments",
			args:  []string{"registry", "update"},
			isErr: true,
		},
	}

	runTestCmd(t, cases)
}
