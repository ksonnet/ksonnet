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
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"path/filepath"
	"sort"

	"github.com/google/go-jsonnet/ast"
	"github.com/ksonnet/ksonnet-lib/ksonnet-gen/astext"
	"github.com/ksonnet/ksonnet-lib/ksonnet-gen/printer"
	"github.com/ksonnet/ksonnet/pkg/app"
	"github.com/ksonnet/ksonnet/pkg/util/jsonnet"
	"github.com/pkg/errors"
)

// Entry is parameter entry. It uses to hold output when listing
// parameters.
type Entry struct {
	// ComponentName is the component that owns this entry.
	ComponentName string
	// ParamName is the name of the parameter.
	ParamName string
	// Value is the value of the parameter.
	Value string
}

// Lister lists parameters.
type Lister struct {
	// Destination is the cluster details.
	Destination app.EnvironmentDestinationSpec
	// AppRoot is the root path for the application.
	AppRoot string

	createEntry func(id string, object *astext.Object) ([]Entry, error)
}

// NewLister creates an instance of Lister.
func NewLister(appRoot string, destination app.EnvironmentDestinationSpec) *Lister {
	ec := newEntryCreator()

	return &Lister{
		AppRoot:     appRoot,
		Destination: destination,

		createEntry: ec.Create,
	}
}

// List lists parameters in a slice of Entry given parameters source in a reader.
func (l *Lister) List(r io.Reader, componentName string) ([]Entry, error) {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, errors.Wrap(err, "reading parameters source")
	}

	source := string(data)

	object, err := l.buildObject(source)
	if err != nil {
		return nil, errors.Wrap(err, "building params object")
	}

	co, err := l.componentsObject(object)
	if err != nil {
		return nil, errors.Wrap(err, "finding parameter components")
	}

	var entries []Entry

	for i := range co.Fields {
		f := co.Fields[i]

		id, err := jsonnet.FieldID(f)
		if err != nil {
			return nil, errors.Wrap(err, "retrieving field from object")
		}

		if componentName != "" && componentName != id {
			// if filter is based on a component name, continue if this component is
			// not in the list
			continue
		}

		paramObject, ok := f.Expr2.(*astext.Object)
		if !ok {
			return nil, errors.Errorf("component %q value is not an object", id)
		}

		paramEntries, err := l.createEntry(id, paramObject)
		if err != nil {
			return nil, err
		}

		entries = append(entries, paramEntries...)
	}

	sort.Slice(entries, func(i, j int) bool {
		if entries[i].ComponentName < entries[j].ComponentName {
			return true
		}

		return entries[i].ParamName < entries[j].ParamName
	})

	return entries, nil
}

// componentsObject finds the components object in side of a Jsonnet object in params format.
func (l *Lister) componentsObject(object *astext.Object) (*astext.Object, error) {
	for i := range object.Fields {
		f := object.Fields[i]

		id, err := jsonnet.FieldID(f)
		if err != nil {
			return nil, errors.Wrap(err, "finding object id")
		}

		if id == "components" {
			co, ok := f.Expr2.(*astext.Object)
			if !ok {
				return nil, errors.Errorf("expected components to be an object, it was %T", f.Expr2)
			}

			return co, nil
		}
	}

	return nil, errors.Errorf("unable to find components object")
}

// buildObject converts params.libsonnet source into a Jsonnet object.
func (l *Lister) buildObject(source string) (*astext.Object, error) {
	// TODO: this code is repeated in module.Resolveparams, and should be centralized.
	envCode, err := l.destinationObject()
	if err != nil {
		return nil, errors.Wrap(err, "building environment object")
	}

	vm := jsonnet.NewVM()
	vm.AddJPath(
		filepath.Join(l.AppRoot, "vendor"),
		filepath.Join(l.AppRoot, "lib"),
	)

	// TODO: move this to pkg/app
	vm.ExtCode("__ksonnet/environments", envCode)

	output, err := vm.EvaluateSnippet("params.libsonnet", source)
	if err != nil {
		return nil, errors.Wrap(err, "evaluating params.libsonnet")
	}

	n, err := jsonnet.ParseNode("params.libsonnet", output)
	if err != nil {
		return nil, errors.Wrap(err, "parsing parameters")
	}

	object, ok := n.(*astext.Object)
	if !ok {
		return nil, errors.Errorf("params.libsonnet did not evaluate to an object (%T)", n)
	}

	return object, nil
}

func (l *Lister) destinationObject() (string, error) {
	dest := map[string]string{
		"server":    l.Destination.Server,
		"namespace": l.Destination.Namespace,
	}

	data, err := json.Marshal(&dest)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// objectValueAsString extracts a value from an object as a string.
func objectValueAsString(source, key string) (string, error) {
	object, err := jsonnet.Parse("objectValueAsString", source)
	if err != nil {
		return "", errors.Wrap(err, "parsing object")
	}

	for i := range object.Fields {
		f := object.Fields[i]
		id, err := jsonnet.FieldID(f)
		if err != nil {
			return "", errors.Wrap(err, "finding id of field")
		}

		if id != key {
			continue
		}

		switch t := f.Expr2.(type) {
		case *astext.Object:
			t.Oneline = true
		case *ast.Array:
			// force array to print on a single line by setting its begin line equal to its
			// end line.
			loc := t.NodeBase.Loc()
			loc.Begin.Line = 1
			loc.End.Line = 1
		}

		var buf bytes.Buffer
		if err = printer.Fprint(&buf, f.Expr2); err != nil {
			return "", errors.Wrap(err, "converting node to text")
		}

		return buf.String(), nil
	}

	return "", errors.Errorf("object did not contain key %q", key)
}

// entryCreator creates Entry from a param object.
type entryCreator struct {
	idField func(astext.ObjectField) (string, error)
}

func newEntryCreator() *entryCreator {
	return &entryCreator{
		idField: jsonnet.FieldID,
	}
}

func (ec *entryCreator) Create(id string, object *astext.Object) ([]Entry, error) {
	var buf bytes.Buffer
	if err := printer.Fprint(&buf, object); err != nil {
		return nil, errors.Wrapf(err, "converting %q object to text")
	}

	var entries []Entry

	for i := range object.Fields {
		paramField := object.Fields[i]

		paramID, err := ec.idField(paramField)
		if err != nil {
			return nil, errors.Wrapf(err, "retrieving parameter field from component %q", id)
		}

		val, err := objectValueAsString(buf.String(), paramID)
		if err != nil {
			return nil, errors.Wrapf(err, "finding value for key %q in component %q", paramID, id)
		}

		e := Entry{
			ComponentName: id,
			ParamName:     paramID,
			Value:         val,
		}

		entries = append(entries, e)
	}

	return entries, nil
}
