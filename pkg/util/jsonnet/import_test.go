package jsonnet

import (
	"testing"

	"github.com/google/go-jsonnet/ast"
	"github.com/ksonnet/ksonnet-lib/ksonnet-gen/astext"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

func stageContent(t *testing.T, fs afero.Fs, path string, data []byte) {
	err := afero.WriteFile(importFs, path, data, 0644)
	require.NoError(t, err)
}

func TestImport(t *testing.T) {
	ogFs := importFs
	defer func(ogFs afero.Fs) {
		importFs = ogFs
	}(ogFs)

	importFs = afero.NewMemMapFs()

	stageContent(t, importFs, "/obj.jsonnet", []byte("{}"))
	stageContent(t, importFs, "/array.jsonnet", []byte(`["a", "b"]`))
	stageContent(t, importFs, "/parser.jsonnet", []byte("local️ a = b; []"))

	cases := []struct {
		name  string
		path  string
		isErr bool
	}{
		{
			name: "with an existing jsonnet file",
			path: "/obj.jsonnet",
		},
		{
			name:  "no filename",
			isErr: true,
		},
		{
			name:  "invalid file",
			path:  "/invalid",
			isErr: true,
		},
		{
			name:  "parser error",
			path:  "/parser.jsonnet",
			isErr: true,
		},
		{
			name:  "not an object",
			path:  "/array.jsonnet",
			isErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			obj, err := Import(tc.path)
			if tc.isErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)

				obj.NodeBase = ast.NodeBase{}
				expected := &astext.Object{}

				require.Equal(t, expected, obj)
			}
		})
	}

}
