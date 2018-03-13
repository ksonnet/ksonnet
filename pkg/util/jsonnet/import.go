package jsonnet

import (
	"github.com/ksonnet/ksonnet-lib/ksonnet-gen/astext"
	"github.com/ksonnet/ksonnet/pkg/docparser"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
)

var (
	// importFs is the filesystem import uses when a importFs is not supplied.
	importFs = afero.NewOsFs()
)

// Import imports jsonnet from a path.
func Import(filename string) (*astext.Object, error) {
	return ImportFromFs(filename, importFs)
}

// ImportFromFs imports jsonnet from a path on an afero filesystem.
func ImportFromFs(filename string, fs afero.Fs) (*astext.Object, error) {
	if filename == "" {
		return nil, errors.New("filename was blank")
	}

	b, err := afero.ReadFile(fs, filename)
	if err != nil {
		return nil, errors.Wrap(err, "read lib")
	}

	return Parse(filename, string(b))

}

// Parse converts a jsonnet snippet to AST.
func Parse(filename, src string) (*astext.Object, error) {
	tokens, err := docparser.Lex(filename, src)
	if err != nil {
		return nil, errors.Wrap(err, "lex jsonnet snippet")
	}

	node, err := docparser.Parse(tokens)
	if err != nil {
		return nil, errors.Wrap(err, "parse jsonnet snippet")
	}

	root, ok := node.(*astext.Object)
	if !ok {
		return nil, errors.New("root was not an object")
	}

	return root, nil
}
