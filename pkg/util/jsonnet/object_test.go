package jsonnet

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/google/go-jsonnet/ast"
	"github.com/ksonnet/ksonnet-lib/ksonnet-gen/astext"
	nm "github.com/ksonnet/ksonnet-lib/ksonnet-gen/nodemaker"
	"github.com/ksonnet/ksonnet-lib/ksonnet-gen/printer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSet(t *testing.T) {
	labels := map[string]interface{}{
		"metadata": map[string]interface{}{
			"labels": map[string]interface{}{
				"label": "label",
			},
		},
	}

	labelsObject, err := nm.KVFromMap(labels)
	require.NoError(t, err)

	cases := []struct {
		name       string
		updatePath []string
		update     ast.Node
		expected   string
		isErr      bool
	}{
		{
			name:       "update existing field",
			updatePath: []string{"a", "b", "c"},
			update:     nm.NewInt(9).Node(),
			expected:   "{\n  a:: {\n    b:: {\n      c:: 9,\n    },\n  },\n}",
		},
		{
			name:       "set map",
			updatePath: []string{"a", "d"},
			update:     labelsObject.Node(),
			expected:   string(testdata(t, "set-map.jsonnet")),
		},
		{
			name:       "set new field",
			updatePath: []string{"a", "e"},
			update:     nm.NewInt(9).Node(),
			expected:   "{\n  a:: {\n    b:: {\n      c:: \"value\",\n    },\n    e: 9,\n  },\n}",
		},
		{
			name:       "set object field to non object",
			updatePath: []string{"a"},
			update:     nm.NewInt(9).Node(),
			isErr:      true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			b := nm.NewObject()
			b.Set(nm.NewKey("c"), nm.NewStringDouble("value"))

			a := nm.NewObject()
			a.Set(nm.NewKey("b"), b)

			object := nm.NewObject()
			object.Set(nm.NewKey("a"), a)

			astObject := object.Node().(*astext.Object)

			err := Set(astObject, tc.updatePath, tc.update)
			if tc.isErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)

				var got bytes.Buffer
				err = printer.Fprint(&got, astObject)
				require.NoError(t, err)

				require.Equal(t, tc.expected, got.String())
			}
		})
	}
}

func TestFindObject(t *testing.T) {
	b := nm.NewObject()
	b.Set(nm.NewKey("c"), nm.NewStringDouble("value"))

	a := nm.NewObject()
	a.Set(nm.NewKey("b"), b)
	a.Set(nm.NewKey("d-1", nm.KeyOptCategory(ast.ObjectFieldStr)), nm.NewStringDouble("string"))

	object := nm.NewObject()
	object.Set(nm.NewKey("a"), a)

	astObject := object.Node().(*astext.Object)

	cases := []struct {
		name     string
		path     []string
		expected ast.Node
		isErr    bool
	}{
		{
			name:     "find nested object",
			path:     []string{"a", "b", "c"},
			expected: b.Node(),
		},
		{
			name:     "find string id object",
			path:     []string{"a", "d-1"},
			expected: a.Node(),
		},
		{
			name:  "invalid path",
			path:  []string{"z"},
			isErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			node, err := FindObject(astObject, tc.path)
			if tc.isErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expected, node)

			}

		})
	}
}

func TestFieldID(t *testing.T) {

	expr1Field := astext.ObjectField{
		ObjectField: ast.ObjectField{
			Expr1: nm.NewStringDouble("my-field").Node(),
		},
	}

	invalidExpr1Field := astext.ObjectField{
		ObjectField: ast.ObjectField{
			Expr1: nm.NewInt(1).Node(),
		},
	}

	id := ast.Identifier("my-field")
	idField := astext.ObjectField{
		ObjectField: ast.ObjectField{
			Id: &id,
		},
	}

	cases := []struct {
		name     string
		field    astext.ObjectField
		expected string
		isErr    bool
	}{
		{
			name:  "no id",
			isErr: true,
		},
		{
			name:     "field with id in Expr1",
			field:    expr1Field,
			expected: "my-field",
		},
		{
			name:  "field with invalid Expr1",
			field: invalidExpr1Field,
			isErr: true,
		},
		{
			name:     "field with id as Identifier",
			field:    idField,
			expected: "my-field",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			id, err := FieldID(tc.field)
			if tc.isErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expected, id)
			}
		})
	}
}

func testdata(t *testing.T, name string) []byte {
	b, err := ioutil.ReadFile("testdata/" + name)
	require.NoError(t, err, "read testdata %s", name)
	return b
}
