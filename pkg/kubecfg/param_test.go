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
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSanitizeParamValue(t *testing.T) {
	tests := []struct {
		value    string
		expected string
	}{
		// numbers
		{
			value:    "64.5",
			expected: "64.5",
		},
		{
			value:    "64",
			expected: "64",
		},
		// boolean
		{
			value:    "false",
			expected: `"false"`,
		},
		// string
		{
			value:    "my string",
			expected: `"my string"`,
		},
	}

	for _, te := range tests {
		got := sanitizeParamValue(te.value)
		require.Equal(t, got, te.expected, "Unexpected sanitized param value")
	}
}
