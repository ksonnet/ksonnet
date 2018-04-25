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

package prototype

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseTemplateType(t *testing.T) {
	cases := []struct {
		name     string
		expected TemplateType
		isErr    bool
	}{
		{
			name:     "yaml",
			expected: YAML,
		},
		{
			name:     "yml",
			expected: YAML,
		},
		{
			name:     "json",
			expected: JSON,
		},
		{
			name:     "jsonnet",
			expected: Jsonnet,
		},
		{
			name:  "unknown",
			isErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tt, err := ParseTemplateType(tc.name)
			if tc.isErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.expected, tt)
		})
	}
}
