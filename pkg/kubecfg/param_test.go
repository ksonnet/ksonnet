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

	param "github.com/ksonnet/ksonnet/metadata/params"

	"github.com/stretchr/testify/require"
)

func TestDiffParams(t *testing.T) {
	tests := []struct {
		params1  map[string]param.Params
		params2  map[string]param.Params
		expected []*paramDiffRecord
	}{
		{
			map[string]param.Params{
				"bar": param.Params{"replicas": "4"},
				"foo": param.Params{"replicas": "1", "name": `"foo"`},
			},
			map[string]param.Params{
				"bar": param.Params{"replicas": "3"},
				"foo": param.Params{"name": `"foo-dev"`, "replicas": "1"},
				"baz": param.Params{"name": `"baz"`, "replicas": "4"},
			},
			[]*paramDiffRecord{
				&paramDiffRecord{
					component: "bar",
					param:     "replicas",
					value1:    "4",
					value2:    "3",
				},
				&paramDiffRecord{
					component: "baz",
					param:     "name",
					value1:    "",
					value2:    `"baz"`,
				},
				&paramDiffRecord{
					component: "baz",
					param:     "replicas",
					value1:    "",
					value2:    "4",
				},
				&paramDiffRecord{
					component: "foo",
					param:     "name",
					value1:    `"foo"`,
					value2:    `"foo-dev"`,
				},
				&paramDiffRecord{
					component: "foo",
					param:     "replicas",
					value1:    "1",
					value2:    "1",
				},
			},
		},
	}

	for _, s := range tests {
		records := diffParams(s.params1, s.params2)
		require.Equal(t, len(records), len(s.expected), "Record lengths not equivalent")
		for i, record := range records {
			require.EqualValues(t, *s.expected[i], *record)
		}
	}
}

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
