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
)

func Test_envCmd(t *testing.T) {
	cases := []cmdTestCase{
		{
			name:  "no command",
			args:  []string{"env"},
			isErr: true,
		},
		{
			name:  "invalid command",
			args:  []string{"env", "invalid"},
			isErr: true,
		},
	}

	runTestCmd(t, cases)
}
