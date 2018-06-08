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

	ksstrings "github.com/ksonnet/ksonnet/pkg/util/strings"
	"github.com/ksonnet/ksonnet/pkg/util/test"
	"github.com/spf13/afero"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/require"
)

func TestValues(t *testing.T) {
	proto := &Prototype{
		Params: ParamSchemas{
			{
				Name: "name",
				Type: String,
			},
			{
				Name: "key1",
				Type: String,
			},
			{
				Name:    "key2",
				Type:    String,
				Default: ksstrings.Ptr("default"),
			},
		},
	}

	cases := []struct {
		name     string
		p        *Prototype
		flags    *pflag.FlagSet
		args     []string
		isErr    bool
		expected map[string]string
	}{
		{
			name: "from flags",
			p:    proto,
			args: []string{
				"--name", "name",
			},
			flags: BindFlags(proto),
			expected: map[string]string{
				"name": `"name"`,
				"key1": `""`,
				"key2": `"default"`,
			},
		},
		{
			name:  "flags: prototype is nil",
			flags: BindFlags(proto),
			args:  []string{},
			isErr: true,
		},
		{
			name:  "flags: missing default parameter",
			flags: BindFlags(proto),
			args:  []string{},
			isErr: true,
		},
		{
			name:  "from values file",
			p:     proto,
			flags: BindFlags(proto),
			args: []string{
				"--name", "name",
				"--values-file", "/files/valid",
			},
			expected: map[string]string{
				"name": `"name"`,
				"key1": `"value1"`,
				"key2": `"default"`,
			},
		},
		{
			name:  "from values file with equal",
			p:     proto,
			flags: BindFlags(proto),
			args: []string{
				"--name", "name",
				"--values-file", "/files/with-equal",
			},
			expected: map[string]string{
				"name": `"name"`,
				"key1": `"value1=value2"`,
				"key2": `"default"`,
			},
		},
		{
			name:  "values: invalid",
			p:     proto,
			flags: BindFlags(proto),
			args: []string{
				"--name", "name",
				"--values-file", "/files/invalid",
			},
			isErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fs := afero.NewMemMapFs()
			test.StageDir(t, fs, "values-files", "/files")

			if tc.flags != nil {
				err := tc.flags.Parse(tc.args)
				require.NoError(t, err)
			}

			got, err := Values(fs, tc.p, tc.flags)
			if tc.isErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			require.Equal(t, tc.expected, got)
		})
	}
}
