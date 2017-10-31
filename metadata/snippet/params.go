// Copyright 2017 The kubecfg authors
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

package snippet

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"github.com/google/go-jsonnet/ast"
	"github.com/google/go-jsonnet/parser"
)

const (
	componentsID = "components"
)

func visitComponentsObj(component, snippet string) (*ast.Node, error) {
	tokens, err := parser.Lex(component, snippet)
	if err != nil {
		return nil, err
	}

	root, err := parser.Parse(tokens)
	if err != nil {
		return nil, err
	}

	switch n := root.(type) {
	case *ast.Object:
		for _, field := range n.Fields {
			if field.Id != nil && *field.Id == componentsID {
				return &field.Expr2, nil
			}
		}
	default:
		return nil, fmt.Errorf("Expected node type to be object")
	}
	// If this point has been reached, it means we weren't able to find a top-level components object.
	return nil, fmt.Errorf("Invalid format; expected to find a top-level components object")
}

func writeParams(params map[string]string) string {
	// keys maintains an alphabetically sorted list of the param keys
	keys := make([]string, 0, len(params))
	for key := range params {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var buffer bytes.Buffer
	buffer.WriteString("\n")
	for i, key := range keys {
		buffer.WriteString(fmt.Sprintf("      %s: %s,", key, params[key]))
		if i < len(keys)-1 {
			buffer.WriteString("\n")
		}
	}
	buffer.WriteString("\n")
	return buffer.String()
}

func appendComponent(component, snippet string, params map[string]string) (string, error) {
	componentsNode, err := visitComponentsObj(component, snippet)
	if err != nil {
		return "", err
	}

	// Find the location to append the next component
	switch n := (*componentsNode).(type) {
	case *ast.Object:
		// Ensure that the component we are trying to create params for does not already exist.
		for _, field := range n.Fields {
			if field.Id != nil && string(*field.Id) == component {
				return "", fmt.Errorf("Component parameters for '%s' already exists", component)
			}
		}
	default:
		return "", fmt.Errorf("Expected components node type to be object")
	}

	lines := strings.Split(snippet, "\n")

	// Get an alphabetically sorted list of the param keys
	keys := make([]string, 0, len(params))
	for key := range params {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	// Create the jsonnet resembling the component params
	var buffer bytes.Buffer
	buffer.WriteString("    " + component + ": {")
	buffer.WriteString(writeParams(params))
	buffer.WriteString("    },")

	// Insert the new component to the end of the list of components
	insertLine := (*componentsNode).Loc().End.Line - 1
	lines = append(lines, "")
	copy(lines[insertLine+1:], lines[insertLine:])
	lines[insertLine] = buffer.String()

	return strings.Join(lines, "\n"), nil
}
