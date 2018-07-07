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

package table

import (
	"bytes"
	"io"
	"io/ioutil"
	"testing"

	"github.com/ksonnet/ksonnet/pkg/util/test"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDetectFormat(t *testing.T) {
	cases := []struct {
		name       string
		formatName string
		expected   Format
		isErr      bool
	}{
		{
			name:       "json",
			formatName: "json",
			expected:   FormatJSON,
		},
		{
			name:       "table",
			formatName: "table",
			expected:   FormatTable,
		},
		{
			name:       "blank string",
			formatName: "",
			expected:   FormatTable,
		},
		{
			name:       "unknown format returns an error",
			formatName: "unknown",
			isErr:      true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := DetectFormat(tc.formatName)
			if tc.isErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.expected, got)
		})
	}
}

func TestTable(t *testing.T) {
	cases := []struct {
		name   string
		format Format
		rw     io.ReadWriter
		output string
		isErr  bool
	}{
		{
			name:   "text output",
			format: FormatTable,
			rw:     &bytes.Buffer{},
			output: "table.txt",
		},
		{
			name:   "JSON format",
			format: FormatJSON,
			rw:     &bytes.Buffer{},
			output: "output.json",
		},
		{
			name:   "unknown format",
			format: Format(99),
			rw:     &bytes.Buffer{},
			output: "table.txt",
		},
		{
			name:  "output is nil",
			isErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			table := New("test", tc.rw)

			table.SetHeader([]string{"name", "version", "Namespace", "SERVER"})
			table.Append([]string{"default", "v1.7.0", "default", "http://default"})
			table.AppendBulk([][]string{
				{"dev", "v1.8.0", "dev", "http://dev"},
				{"east/prod", "v1.8.0", "east/prod", "http://east-prod"},
			})

			table.SetFormat(tc.format)

			err := table.Render()
			if tc.isErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			got, err := ioutil.ReadAll(tc.rw)
			require.NoError(t, err)

			expected := test.ReadTestData(t, tc.output)
			assert.Equal(t, expected, string(got))

		})
	}

}

func TestTable_no_header(t *testing.T) {
	cases := []struct {
		name   string
		format Format
		output string
		isErr  bool
	}{
		{
			name:   "in table format",
			format: FormatTable,
			output: "table_no_header.txt",
		},
		{
			name:   "in JSON format",
			format: FormatJSON,
			isErr:  true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			table := New("test", &buf)
			table.SetFormat(tc.format)

			table.Append([]string{"default", "v1.7.0", "default", "http://default"})
			table.AppendBulk([][]string{
				{"dev", "v1.8.0", "dev", "http://dev"},
				{"east/prod", "v1.8.0", "east/prod", "http://east-prod"},
			})

			err := table.Render()
			if tc.isErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			expected := test.ReadTestData(t, tc.output)
			assert.Equal(t, expected, buf.String())

		})
	}
}

func TestTable_header_and_row_length_must_match(t *testing.T) {
	cases := []struct {
		name   string
		format Format
	}{
		{
			name:   "in table format",
			format: FormatTable,
		},
		{
			name:   "in JSON format",
			format: FormatJSON,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			table := New("test", &buf)
			table.SetFormat(tc.format)

			table.SetHeader([]string{"name", "version", "Namespace", "SERVER"})
			table.Append([]string{"default", "v1.7.0", "default"})

			err := table.Render()
			require.Error(t, err)
		})
	}
}

func TestTable_trim_space(t *testing.T) {
	var buf bytes.Buffer
	table := New("test", &buf)

	table.SetHeader([]string{"name", "version", "Namespace", "SERVER"})
	table.Append([]string{"default", "v1.7.0", "default", "http://default"})
	table.AppendBulk([][]string{
		{"dev", "v1.8.0", "", ""},
		{"east/prod", "v1.8.0", "east/prod", "http://east-prod"},
	})

	err := table.Render()
	require.NoError(t, err)

	expected := test.ReadTestData(t, "table_trim_space.txt")
	assert.Equal(t, expected, buf.String())

}

func TestTable_printf_failure(t *testing.T) {
	var buf bytes.Buffer
	table := New("test", &buf)
	table.printf = func(_ io.Writer, _ string, _ ...interface{}) (int, error) {
		return 0, errors.New("printf failure")
	}

	table.SetHeader([]string{"name", "version", "Namespace", "SERVER"})
	table.Append([]string{"default", "v1.7.0", "default", "http://default"})

	err := table.Render()
	require.Error(t, err)
}
