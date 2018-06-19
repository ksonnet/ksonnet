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

func Test_envSetCmd(t *testing.T) {
	cases := []cmdTestCase{
		{
			name:   "in general",
			args:   []string{"env", "set", "default", "--name", "new-name", "--namespace", "new-namespace", "--server", "new-server", "--api-spec", "new-api-spec"},
			action: actionEnvSet,
			expected: map[string]interface{}{
				actions.OptionApp:        nil,
				actions.OptionEnvName:    "default",
				actions.OptionNewEnvName: "new-name",
				actions.OptionNamespace:  "new-namespace",
				actions.OptionServer:     "new-server",
				actions.OptionSpecFlag:   "new-api-spec",
			},
		},
		{
			name:  "no environment",
			args:  []string{"env", "set"},
			isErr: true,
		},
	}

	runTestCmd(t, cases)
}
