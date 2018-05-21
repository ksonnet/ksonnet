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
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/ksonnet/ksonnet/pkg/util/strings"
	"github.com/stretchr/testify/require"
)

func TestJsonnetParse(t *testing.T) {
	cases := []struct {
		name  string
		file  string
		isErr bool
	}{
		{
			name: "full example",
			file: "prototype1",
		},
		{
			name:  "invalid directive",
			file:  "prototype2",
			isErr: true,
		},
		{
			name: "parse comments",
			file: "prototype3",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			b, err := ioutil.ReadFile(filepath.Join("testdata", tc.file+".jsonnet"))
			require.NoError(t, err)

			got, err := JsonnetParse(string(b))
			if tc.isErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			var s Prototype
			expectedJSON := filepath.Join("testdata", tc.file+".json")
			f, err := os.Open(expectedJSON)
			require.NoError(t, err)

			err = json.NewDecoder(f).Decode(&s)
			require.NoError(t, err)

			require.Equal(t, &s, got)
		})
	}
}

func Test_newDirective(t *testing.T) {
	cases := []struct {
		name  string
		src   string
		isErr bool
	}{
		{
			name:  "incomplete prototype directive",
			src:   "incomplete",
			isErr: true,
		},
		{
			name:  "unknown prototype directive",
			src:   "unknown invalid",
			isErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			d := newDirective(tc.src)

			s := &Prototype{}
			err := d(s)
			if tc.isErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}

func Test_paramDirective(t *testing.T) {
	cases := []struct {
		name     string
		src      string
		expected ParamSchemas
		isErr    bool
	}{
		{
			name: "valid directive",
			src:  "name string Name of the service",
			expected: ParamSchemas{
				{
					Name:        "name",
					Alias:       strings.Ptr("name"),
					Description: "Name of the service",
					Type:        String,
				},
			},
		},
		{
			name:  "invalid type",
			src:   "name invalid Name of the service",
			isErr: true,
		},
		{
			name:  "invalid fields",
			src:   "name",
			isErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := &Prototype{}
			fn := paramDirective(tc.src)

			err := fn(s)
			if tc.isErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.expected, s.Params)
		})
	}
}

func Test_optParamDirective(t *testing.T) {
	cases := []struct {
		name     string
		src      string
		expected ParamSchemas
		isErr    bool
	}{
		{
			name: "valid directive",
			src:  "name string name Name of the service",
			expected: ParamSchemas{
				{
					Name:        "name",
					Alias:       strings.Ptr("name"),
					Description: "Name of the service",
					Default:     strings.Ptr("name"),
					Type:        String,
				},
			},
		},
		{
			name:  "invalid type",
			src:   "name invalid Name of the service",
			isErr: true,
		},
		{
			name:  "invalid fields",
			src:   "name",
			isErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := &Prototype{}
			fn := optParamDirective(tc.src)

			err := fn(s)
			if tc.isErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.expected, s.Params)
		})
	}
}
