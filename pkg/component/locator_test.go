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

package component

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_locateTarget(t *testing.T) {
	cases := []struct {
		name     string
		in       string
		expected string
	}{
		{
			name:     "root path",
			in:       ".",
			expected: filepath.ToSlash("/components"),
		},
		{
			name:     "nested path 1",
			in:       "foo",
			expected: filepath.ToSlash("/components/foo"),
		},
		{
			name:     "nested path 2",
			in:       "foo.bar",
			expected: filepath.ToSlash("/components/foo/bar"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := locateTarget(filepath.FromSlash("/"), tc.in)
			require.Equal(t, tc.expected, got)
		})
	}
}
