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
	"strconv"
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
	}
	// If this point has been reached, it means we weren't able to find a top-level components object.
	return nil, fmt.Errorf("Invalid format; expected to find a top-level components object")
}

func visitComponentParams(component ast.Node) (map[string]string, *ast.LocationRange, error) {
	params := make(map[string]string)
	var loc *ast.LocationRange

	switch n := component.(type) {
	case *ast.Object:
		loc = n.Loc()
		for _, field := range n.Fields {
			if field.Id != nil {
				key := string(*field.Id)
				val, err := visitParamValue(field.Expr2)
				if err != nil {
					return nil, nil, err
				}
				params[key] = val
			}
		}
	default:
		return nil, nil, fmt.Errorf("Expected component node type to be object")
	}

	return params, loc, nil
}

// visitParamValue returns a string representation of the param value, quoted
// where necessary. Currently only handles trivial types, ex: string, int, bool
func visitParamValue(param ast.Node) (string, error) {
	switch n := param.(type) {
	case *ast.LiteralNumber:
		return strconv.FormatFloat(n.Value, 'f', -1, 64), nil
	case *ast.LiteralBoolean:
		return strconv.FormatBool(n.Value), nil
	case *ast.LiteralString:
		return fmt.Sprintf(`"%s"`, n.Value), nil
	default:
		return "", fmt.Errorf("Found an unsupported param value type: %T", n)
	}
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

func setComponentParams(component, snippet string, params map[string]string) (string, error) {
	componentsNode, err := visitComponentsObj(component, snippet)
	if err != nil {
		return "", err
	}

	var loc *ast.LocationRange
	var currentParams map[string]string

	switch n := (*componentsNode).(type) {
	case *ast.Object:
		for _, field := range n.Fields {
			if field.Id != nil && string(*field.Id) == component {
				currentParams, loc, err = visitComponentParams(field.Expr2)
				if err != nil {
					return "", err
				}
			}
		}
	default:
		return "", fmt.Errorf("Expected component node type to be object")
	}

	if loc == nil {
		return "", fmt.Errorf("Could not find component identifier '%s' when attempting to set params", component)
	}

	for k, v := range currentParams {
		if _, ok := params[k]; !ok {
			params[k] = v
		}
	}

	// Replace the component param fields
	lines := strings.Split(snippet, "\n")
	paramsSnippet := writeParams(params)
	newSnippet := strings.Join(lines[:loc.Begin.Line], "\n") + paramsSnippet + strings.Join(lines[loc.End.Line-1:], "\n")
	//newSnippet := append(lines[:loc.Begin.Line], paramsSnippet)
	//newSnippet = append(newSnippet, strings.Join(lines[loc.End.Line-1:], "\n"))

	return newSnippet, nil
}
