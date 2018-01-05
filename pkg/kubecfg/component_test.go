// Copyright 2017 The ksonnet authors
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
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPrintComponents(t *testing.T) {
	for _, tc := range []struct {
		components []string
		expected   string
	}{
		{
			components: []string{"a", "b"},
			expected: `COMPONENT
=========
a
b
`,
		},
		// Check that components are displayed in alphabetical order
		{
			components: []string{"b", "a"},
			expected: `COMPONENT
=========
a
b
`,
		},
		// Check empty components scenario
		{
			components: []string{},
			expected: `COMPONENT
=========
`,
		},
	} {
		out, err := printComponents(os.Stdout, tc.components)
		if err != nil {
			t.Fatalf(err.Error())
		}
		require.EqualValues(t, tc.expected, out)
	}
}
