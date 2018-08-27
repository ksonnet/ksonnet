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

func Test_importCmd(t *testing.T) {
	cases := []cmdTestCase{
		{
			name:   "import location without module",
			args:   []string{"import", "-f", "location"},
			action: actionImport,
			expected: map[string]interface{}{
				actions.OptionApp:    nil,
				actions.OptionPath:   "location",
				actions.OptionModule: "/",
			},
		},
		{
			name:   "import location with module",
			args:   []string{"import", "-f", "location", "--module", "module"},
			action: actionImport,
			expected: map[string]interface{}{
				actions.OptionApp:    nil,
				actions.OptionPath:   "location",
				actions.OptionModule: "module",
			},
		},
	}

	runTestCmd(t, cases)
}
