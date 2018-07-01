// Copyright 2018 The kubecfg authors
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

package jsonnet

import (
	"fmt"
	"os"
	"path"

	jsonnet "github.com/google/go-jsonnet"
	"github.com/google/go-jsonnet/ast"
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

// ImportFromFs imports jsonnet object from a path on an afero filesystem.
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

// ImportNodeFromFs imports jsonnet node from a path on an afero filesystem.
func ImportNodeFromFs(filename string, fs afero.Fs) (ast.Node, error) {
	if filename == "" {
		return nil, errors.New("filename was blank")
	}

	b, err := afero.ReadFile(fs, filename)
	if err != nil {
		return nil, errors.Wrap(err, "read lib")
	}

	return ParseNode(filename, string(b))
}

// Parse converts a jsonnet snippet to AST Object.
func Parse(filename, src string) (*astext.Object, error) {
	node, err := ParseNode(filename, src)
	if err != nil {
		return nil, err
	}

	root, ok := node.(*astext.Object)
	if !ok {
		return nil, errors.New("root was not an object")
	}

	return root, nil
}

// ParseNode converts a jsonnet snippet to AST node.
func ParseNode(filename, src string) (ast.Node, error) {
	tokens, err := docparser.Lex(filename, src)
	if err != nil {
		return nil, errors.Wrap(err, "lex jsonnet snippet")
	}

	node, err := docparser.Parse(tokens)
	if err != nil {
		return nil, errors.Wrap(err, "parse jsonnet snippet")
	}

	return node, nil
}

// Importer extends jsonnet.Importer to support setting import paths.
type Importer interface {
	jsonnet.Importer

	AddJPath(paths ...string)
}

// FileImporter extends jsonnet.FileImporter to allow incrementally adding import paths.
type FileImporter struct {
	jsonnet.FileImporter
}

// AddJPath adds the provided paths to the importer.
func (f *FileImporter) AddJPath(paths ...string) {
	f.JPaths = append(f.JPaths, paths...)
}

// AferoImporter implements Importer using an afero Fs interface.
type AferoImporter struct {
	FileImporter
	Fs afero.Fs
}

func (ai *AferoImporter) tryPath(dir, importedPath string) (found bool, content []byte, foundHere string, err error) {
	var absPath string
	if path.IsAbs(importedPath) {
		absPath = importedPath
	} else {
		absPath = path.Join(dir, importedPath)
	}
	content, err = afero.ReadFile(ai.Fs, absPath)
	if os.IsNotExist(err) {
		return false, nil, "", nil
	}
	return true, content, absPath, err
}

// Import imports a file.
func (ai *AferoImporter) Import(dir, importedPath string) (contents jsonnet.Contents, foundHere string, err error) {
	found, content, foundHere, err := ai.tryPath(dir, importedPath)
	if err != nil {
		return jsonnet.MakeContents(""), "", err
	}

	for i := len(ai.JPaths) - 1; !found && i >= 0; i-- {
		found, content, foundHere, err = ai.tryPath(ai.JPaths[i], importedPath)
		if err != nil {
			return jsonnet.MakeContents(""), "", err
		}
	}

	if !found {
		return jsonnet.MakeContents(""), "", fmt.Errorf("couldn't open import %#v: no match locally or in the Jsonnet library paths", importedPath)
	}
	return jsonnet.MakeContents(string(content)), foundHere, nil
}

// ImporterOpt configures a VM with a jsonnet.Importer
func ImporterOpt(importer Importer) VMOpt {
	return func(vm *VM) {
		vm.importer = importer
	}
}

// AferoImporterOpt configures a VM with a jsonnet.Importer
func AferoImporterOpt(fs afero.Fs) VMOpt {
	return func(vm *VM) {
		vm.importer = &AferoImporter{Fs: fs}
	}
}
