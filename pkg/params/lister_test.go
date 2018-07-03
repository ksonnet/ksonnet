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

package params

import (
	"strings"
	"testing"

	"github.com/google/go-jsonnet/ast"
	"github.com/ksonnet/ksonnet-lib/ksonnet-gen/astext"
	"github.com/ksonnet/ksonnet-lib/ksonnet-gen/nodemaker"
	"github.com/ksonnet/ksonnet/pkg/app"
	"github.com/ksonnet/ksonnet/pkg/util/test"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLister_List(t *testing.T) {
	cases := []struct {
		name          string
		init          func(t *testing.T, l *Lister)
		componentName string
		expected      []Entry
		isErr         bool
	}{
		{
			name: "all components",
			expected: []Entry{
				{
					ComponentName: "nginx",
					ParamName:     "a",
					Value:         "80",
				},
				{
					ComponentName: "other-name",
					ParamName:     "b",
					Value:         `'string'`,
				},
			},
		},
		{
			name:          "all components",
			componentName: "other-name",
			expected: []Entry{
				{
					ComponentName: "other-name",
					ParamName:     "b",
					Value:         `'string'`,
				},
			},
		},
		{
			name: "create entry failure",
			init: func(t *testing.T, l *Lister) {
				l.createEntry = func(string, *astext.Object) ([]Entry, error) {
					return nil, errors.New("fail")
				}
			},
			isErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dest := app.EnvironmentDestinationSpec{
				Server:    "https://example.com",
				Namespace: "default",
			}

			l := NewLister("/app", dest)

			if tc.init != nil {
				tc.init(t, l)
			}

			source := test.ReadTestData(t, "lister-params.libsonnet")
			r := strings.NewReader(source)

			got, err := l.List(r, tc.componentName)
			if tc.isErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.expected, got)
		})
	}
}

func Test_objectValueAsString(t *testing.T) {
	cases := []struct {
		name     string
		source   string
		key      string
		expected string
		isErr    bool
	}{
		{
			name:     "string",
			source:   `{foo: "bar"}`,
			key:      "foo",
			expected: `'bar'`,
		},
		{
			name:     "int",
			source:   `{foo: 9}`,
			key:      "foo",
			expected: `9`,
		},
		{
			name:     "float",
			source:   `{foo: 9.9}`,
			key:      "foo",
			expected: `9.9`,
		},
		{
			name:     "float",
			source:   `{foo: [9.9]}`,
			key:      "foo",
			expected: "[9.9]",
		},
		{
			name:     "object",
			source:   "{foo: {bar: 9.9}}",
			key:      "foo",
			expected: "{ bar: 9.9 }",
		},
		{
			name:   "source is not jsonnet object",
			source: `-`,
			key:    "foo",
			isErr:  true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := objectValueAsString(tc.source, tc.key)
			if tc.isErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.expected, got)
		})
	}

}

func Test_entryCreator_Create(t *testing.T) {
	ec := newEntryCreator()

	id := ast.Identifier("id")
	o := &astext.Object{
		Fields: []astext.ObjectField{
			{
				ObjectField: ast.ObjectField{
					Id:    &id,
					Kind:  ast.ObjectFieldID,
					Hide:  ast.ObjectFieldInherit,
					Expr2: nodemaker.NewStringDouble("value").Node(),
				},
			},
		},
	}

	got, err := ec.Create("nginx", o)
	require.NoError(t, err)

	expected := []Entry{{ComponentName: "nginx", ParamName: "id", Value: `'value'`}}

	assert.Equal(t, expected, got)
}

func Test_entryCreator_Create_invalid_field(t *testing.T) {
	ec := newEntryCreator()
	ec.idField = func(astext.ObjectField) (string, error) {
		return "", errors.New("failed")
	}

	id := ast.Identifier("id")
	o := &astext.Object{
		Fields: []astext.ObjectField{
			{
				ObjectField: ast.ObjectField{
					Id:    &id,
					Expr2: nodemaker.NewStringDouble("value").Node(),
				},
			},
		},
	}

	_, err := ec.Create("nginx", o)
	require.Error(t, err)
}
