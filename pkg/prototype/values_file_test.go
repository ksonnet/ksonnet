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
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValuesFile_Keys(t *testing.T) {
	cases := []struct {
		name     string
		src      string
		expected []string
		isErr    bool
	}{
		{
			name:     "valid",
			src:      validValuesFile,
			expected: []string{"arr", "name", "num", "obj"},
		},
		{
			name:  "invalid",
			src:   invalidValuesFile,
			isErr: true,
		},
		{
			name:  "invalid field id",
			src:   invalidKeyValuesFile,
			isErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			vf := NewValuesFile(tc.src)
			got, err := vf.Keys()
			if tc.isErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.expected, got)
		})
	}
}

func TestValuesFile_Get(t *testing.T) {
	cases := []struct {
		name     string
		key      string
		expected string
		isErr    bool
	}{
		{
			name:     "retrieve string",
			key:      "name",
			expected: `"name"`,
		},
		{
			name:     "retrieve object",
			key:      "obj",
			expected: "{\n   \"k\": \"v\"\n}",
		},
		{
			name:     "retrieve number",
			key:      "num",
			expected: `9`,
		},
		{
			name:     "retrieve array",
			key:      "arr",
			expected: "[\n   1,\n   2,\n   3\n]",
		},
		{
			name:  "invalid key",
			key:   "invalid",
			isErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			vf := NewValuesFile(validValuesFile)
			got, err := vf.Get(tc.key)
			if tc.isErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			assert.Equal(t, tc.expected, got)
		})
	}
}

func TestReadValues(t *testing.T) {
	cases := []struct {
		name  string
		r     io.Reader
		isErr bool
	}{
		{
			name: "valid jsonnet",
			r:    strings.NewReader(validValuesFile),
		},
		{
			name:  "blank jsonnet",
			r:     strings.NewReader(""),
			isErr: true,
		},
		{
			name:  "nil reader",
			isErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			vf, err := ReadValues(tc.r)
			if tc.isErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			assert.NotEmpty(t, vf.src)
		})
	}
}

var validValuesFile = `
{
	name: "name",
	obj: {k: "v"},
	num: 9,
	arr: [1,2,3],
}
`

var invalidValuesFile = `
{`

var invalidKeyValuesFile = `
{
	[x]: "name",
}`
