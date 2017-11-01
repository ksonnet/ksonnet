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

package params

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

func astRoot(component, snippet string) (ast.Node, error) {
	tokens, err := parser.Lex(component, snippet)
	if err != nil {
		return nil, err
	}

	return parser.Parse(tokens)
}

func visitParams(component ast.Node) (Params, *ast.LocationRange, error) {
	params := make(Params)
	var loc *ast.LocationRange

	n, isObj := component.(*ast.Object)
	if !isObj {
		return nil, nil, fmt.Errorf("Expected component node type to be object")
	}

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

	return params, loc, nil
}

func visitAllParams(components ast.Object) (map[string]Params, error) {
	params := make(map[string]Params)

	for _, f := range components.Fields {
		p, _, err := visitParams(f.Expr2)
		if err != nil {
			return nil, err
		}
		params[string(*f.Id)] = p
	}

	return params, nil
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
		return "", fmt.Errorf("Found an unsupported param AST node type: %T", n)
	}
}

func writeParams(indent int, params Params) string {
	// keys maintains an alphabetically sorted list of the param keys
	keys := make([]string, 0, len(params))
	for key := range params {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var indentBuffer bytes.Buffer
	for i := 0; i < indent; i++ {
		indentBuffer.WriteByte(' ')
	}

	var buffer bytes.Buffer
	buffer.WriteString("\n")
	for i, key := range keys {
		buffer.WriteString(fmt.Sprintf("%s%s: %s,", indentBuffer.String(), key, params[key]))
		if i < len(keys)-1 {
			buffer.WriteString("\n")
		}
	}
	buffer.WriteString("\n")
	return buffer.String()
}

// ---------------------------------------------------------------------------
// Component Parameter-specific functionality

func visitComponentsObj(component, snippet string) (*ast.Object, error) {
	root, err := astRoot(component, snippet)
	if err != nil {
		return nil, err
	}

	n, isObj := root.(*ast.Object)
	if !isObj {
		return nil, fmt.Errorf("Invalid format; expected to find a top-level object")
	}

	for _, field := range n.Fields {
		if field.Id != nil && *field.Id == componentsID {
			c, isObj := field.Expr2.(*ast.Object)
			if !isObj {
				return nil, fmt.Errorf("Expected components node type to be object")
			}
			return c, nil
		}
	}
	// If this point has been reached, it means we weren't able to find a top-level components object.
	return nil, fmt.Errorf("Invalid format; expected to find a top-level components object")
}

func appendComponent(component, snippet string, params Params) (string, error) {
	componentsNode, err := visitComponentsObj(component, snippet)
	if err != nil {
		return "", err
	}

	// Ensure that the component we are trying to create params for does not already exist.
	for _, field := range componentsNode.Fields {
		if field.Id != nil && string(*field.Id) == component {
			return "", fmt.Errorf("Component parameters for '%s' already exists", component)
		}
	}

	lines := strings.Split(snippet, "\n")

	// Create the jsonnet resembling the component params
	var buffer bytes.Buffer
	buffer.WriteString("    " + component + ": {")
	buffer.WriteString(writeParams(6, params))
	buffer.WriteString("    },")

	// Insert the new component to the end of the list of components
	insertLine := (*componentsNode).Loc().End.Line - 1
	lines = append(lines, "")
	copy(lines[insertLine+1:], lines[insertLine:])
	lines[insertLine] = buffer.String()

	return strings.Join(lines, "\n"), nil
}

func getComponentParams(component, snippet string) (Params, *ast.LocationRange, error) {
	componentsNode, err := visitComponentsObj(component, snippet)
	if err != nil {
		return nil, nil, err
	}

	for _, field := range componentsNode.Fields {
		if field.Id != nil && string(*field.Id) == component {
			return visitParams(field.Expr2)
		}
	}

	return nil, nil, fmt.Errorf("Could not find component identifier '%s' when attempting to set params", component)
}

func getAllComponentParams(snippet string) (map[string]Params, error) {
	componentsNode, err := visitComponentsObj("", snippet)
	if err != nil {
		return nil, err
	}

	return visitAllParams(*componentsNode)
}

func setComponentParams(component, snippet string, params Params) (string, error) {
	currentParams, loc, err := getComponentParams(component, snippet)
	if err != nil {
		return "", err
	}

	for k, v := range currentParams {
		if _, ok := params[k]; !ok {
			params[k] = v
		}
	}

	// Replace the component param fields
	lines := strings.Split(snippet, "\n")
	paramsSnippet := writeParams(6, params)
	newSnippet := strings.Join(lines[:loc.Begin.Line], "\n") + paramsSnippet + strings.Join(lines[loc.End.Line-1:], "\n")

	return newSnippet, nil
}

// ---------------------------------------------------------------------------
// Environment Parameter-specific functionality

func findEnvComponentsObj(node ast.Node) (*ast.Object, error) {
	switch n := node.(type) {
	case *ast.Local:
		return findEnvComponentsObj(n.Body)
	case *ast.Binary:
		return findEnvComponentsObj(n.Right)
	case *ast.Object:
		for _, f := range n.Fields {
			if *f.Id == "components" {
				c, isObj := f.Expr2.(*ast.Object)
				if !isObj {
					return nil, fmt.Errorf("Expected components node type to be object")
				}
				return c, nil
			}
		}
		return nil, fmt.Errorf("Invalid params schema -- found %T that is not 'components'", n)
	}
	return nil, fmt.Errorf("Invalid params schema -- did not expect type: %T", node)
}

func getEnvironmentParams(component, snippet string) (Params, *ast.LocationRange, bool, error) {
	root, err := astRoot(component, snippet)
	if err != nil {
		return nil, nil, false, err
	}

	n, err := findEnvComponentsObj(root)
	if err != nil {
		return nil, nil, false, err
	}

	for _, f := range n.Fields {
		if f.Id != nil && string(*f.Id) == component {
			params, loc, err := visitParams(f.Expr2)
			return params, loc, true, err
		}
	}
	// If this point has been reached, it's because we don't have the
	// component in the list of params, return the location after the
	// last field of the components obj
	loc := ast.LocationRange{
		Begin: ast.Location{Line: n.Loc().End.Line - 1, Column: n.Loc().End.Column},
		End:   ast.Location{Line: n.Loc().End.Line, Column: n.Loc().End.Column},
	}

	return make(Params), &loc, false, nil
}

func getAllEnvironmentParams(snippet string) (map[string]Params, error) {
	root, err := astRoot("", snippet)
	if err != nil {
		return nil, err
	}

	componentsNode, err := findEnvComponentsObj(root)
	if err != nil {
		return nil, err
	}

	return visitAllParams(*componentsNode)
}

func setEnvironmentParams(component, snippet string, params Params) (string, error) {
	currentParams, loc, hasComponent, err := getEnvironmentParams(component, snippet)
	if err != nil {
		return "", err
	}

	for k, v := range currentParams {
		if _, ok := params[k]; !ok {
			params[k] = v
		}
	}

	// Replace the component param fields
	var paramsSnippet string
	lines := strings.Split(snippet, "\n")
	if !hasComponent {
		var buffer bytes.Buffer
		buffer.WriteString(fmt.Sprintf("\n    %s +: {", component))
		buffer.WriteString(writeParams(6, params))
		buffer.WriteString("    },\n")
		paramsSnippet = buffer.String()
	} else {
		paramsSnippet = writeParams(6, params)
	}
	newSnippet := strings.Join(lines[:loc.Begin.Line], "\n") + paramsSnippet + strings.Join(lines[loc.End.Line-1:], "\n")

	return newSnippet, nil
}
