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

	"github.com/stretchr/testify/require"
)

func Test_parseParamType(t *testing.T) {
	cases := []struct {
		name     string
		expected ParamType
		isErr    bool
	}{
		{
			name:     "number",
			expected: Number,
		},
		{
			name:     "string",
			expected: String,
		},
		{
			name:     "numberOrString",
			expected: NumberOrString,
		},
		{
			name:     "object",
			expected: Object,
		},
		{
			name:     "array",
			expected: Array,
		},
		{
			name:  "invalid",
			isErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseParamType(tc.name)
			if tc.isErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			require.Equal(t, tc.expected, got)
		})
	}
}

func TestParseParam_String(t *testing.T) {
	cases := []struct {
		name     string
		in       ParamType
		expected string
	}{
		{
			name: "number",
			in:   Number,
		},
		{
			name: "string",
			in:   String,
		},
		{
			name: "numberOrString",
			in:   NumberOrString,
		},
		{
			name: "object",
			in:   Object,
		},
		{
			name: "array",
			in:   Array,
		},
		{
			name: "unknown",
			in:   ParamType("unknown"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := tc.in.String()
			require.Equal(t, tc.name, s)
		})
	}
}
