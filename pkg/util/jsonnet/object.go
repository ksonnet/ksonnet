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
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/google/go-jsonnet/ast"
	"github.com/ksonnet/ksonnet-lib/ksonnet-gen/astext"
	"github.com/pkg/errors"
)

// Set sets an object key at path to a value.
func Set(object *astext.Object, path []string, value ast.Node) error {
	if len(path) == 0 {
		return errors.New("path was empty")
	}

	curObj := object

	for i, k := range path {
		field, err := findField(curObj, k)
		if err != nil {
			switch err.(type) {
			default:
				return err
			case *unknownField:
				field, err = astext.CreateField(k)
				if err != nil {
					return err
				}
				field.Hide = ast.ObjectFieldInherit
				curObj.Fields = append(curObj.Fields, *field)
			}
		}

		if i == len(path)-1 {
			field, _ = findField(curObj, k)
			if canUpdateObject(field.Expr2, value) {
				return errors.New("can't set object to non object")
			}
			field.Expr2 = value
			return nil
		}

		if field.Expr2 == nil {
			curObj = &astext.Object{}
			field.Expr2 = curObj
		} else if obj, ok := field.Expr2.(*astext.Object); ok {
			curObj = obj
		} else {
			return errors.Errorf("child is not an object at %q", k)
		}
	}

	return nil
}

func canUpdateObject(node1, node2 ast.Node) bool {
	return isNodeObject(node1) && !isNodeObject(node2)
}

func isNodeObject(node ast.Node) bool {
	_, ok := node.(*astext.Object)
	return ok
}

type unknownField struct {
	name string
}

func (e *unknownField) Error() string {
	return fmt.Sprintf("unable to find field %q", e.name)
}

func findField(object *astext.Object, id string) (*astext.ObjectField, error) {
	for i := range object.Fields {
		fieldID, err := FieldID(object.Fields[i])
		if err != nil {
			return nil, err
		}

		if id == fieldID {
			return &object.Fields[i], nil
		}
	}

	return nil, &unknownField{name: id}
}

// FindObject finds a path in an object.
func FindObject(object *astext.Object, path []string) (*astext.Object, error) {
	if len(path) == 0 {
		return nil, errors.New("search path was empty")
	}

	for i := range object.Fields {
		id, err := FieldID(object.Fields[i])
		if err != nil {
			return nil, err
		}

		if path[0] == id {
			if len(path) == 1 {

				return object, nil
			}

			child, ok := object.Fields[i].Expr2.(*astext.Object)
			if !ok {
				return nil, errors.Errorf("child is a %T. expected an object", object.Fields[i].Expr2)
			}

			return FindObject(child, path[1:])
		}
	}

	return nil, errors.Errorf("path %s was not found", strings.Join(path, "."))
}

// HasField returns true if the object has a field.
func HasField(object *astext.Object, name string) bool {
	for _, f := range object.Fields {
		id, err := FieldID(f)
		if err != nil {
			return false
		}

		if id == name {
			return true
		}
	}

	return false
}

// FieldID returns the id for an object field.
func FieldID(field astext.ObjectField) (string, error) {
	if field.Expr1 != nil {
		lf, ok := field.Expr1.(*ast.LiteralString)
		if !ok {
			return "", errors.New("field Expr1 is not a string")
		}

		return lf.Value, nil
	}

	if field.Id == nil {
		return "", errors.New("field does not have an ID")
	}

	return string(*field.Id), nil
}

// ConvertObjectToMap converts an object to a map.
func ConvertObjectToMap(obj *astext.Object) (map[string]interface{}, error) {
	m := make(map[string]interface{})

	for i := range obj.Fields {
		id, err := FieldID(obj.Fields[i])
		if err != nil {
			return nil, err
		}

		switch t := obj.Fields[i].Expr2.(type) {
		default:
			return nil, errors.Errorf("unknown value type %T", t)
		case *ast.LiteralString, *ast.LiteralBoolean, *ast.LiteralNumber, *ast.LiteralNull:
			v, err := nodeValue(t)
			if err != nil {
				return nil, err
			}
			m[id] = v
		case *ast.Array:
			array, err := arrayValues(t)
			if err != nil {
				return nil, err
			}
			m[id] = array
		case *astext.Object:
			child, err := ConvertObjectToMap(t)
			if err != nil {
				return nil, err
			}

			m[id] = child
		}

	}

	return m, nil
}

func nodeValue(node ast.Node) (interface{}, error) {
	switch t := node.(type) {
	default:
		return nil, errors.Errorf("unknown value type %T", t)
	case *ast.LiteralString:
		return t.Value, nil
	case *ast.LiteralBoolean:
		return t.Value, nil
	case *ast.LiteralNumber:
		return DecodeValue(fmt.Sprint(t.Value))
	case *ast.LiteralNull:
		return nil, nil
	}
}

func arrayValues(array *ast.Array) ([]interface{}, error) {
	out := make([]interface{}, 0)
	for i := range array.Elements {
		switch t := array.Elements[i].(type) {
		default:
			v, err := nodeValue(array.Elements[i])
			if err != nil {
				return nil, errors.Errorf("arrays can't contain %T", array.Elements[i])
			}
			out = append(out, v)
		case *astext.Object:
			child, err := ConvertObjectToMap(t)
			if err != nil {
				return nil, err
			}

			out = append(out, child)
		}

	}

	return out, nil
}

var (
	reFloat = regexp.MustCompile(`^[-+]?[0-9]*\.?[0-9]+$`)
	reInt   = regexp.MustCompile(`^([+-]?[1-9]\d*|0)$`)
	reArray = regexp.MustCompile(`^\[`)
	reMap   = regexp.MustCompile(`^\{`)
)

// DecodeValue decodes a string to an interface value.
// nolint: gocyclo
func DecodeValue(s string) (interface{}, error) {
	if s == "" {
		return nil, errors.New("value was blank")
	}

	switch {
	case reInt.MatchString(s):
		return strconv.Atoi(s)
	case reFloat.MatchString(s):
		return strconv.ParseFloat(s, 64)
	case strings.ToLower(s) == "true" || strings.ToLower(s) == "false":
		return strconv.ParseBool(s)
	case reArray.MatchString(s):
		var array []interface{}
		if err := json.Unmarshal([]byte(s), &array); err != nil {
			return nil, errors.Errorf("array value is badly formatted: %s", s)
		}
		return array, nil
	case reMap.MatchString(s):
		var obj map[string]interface{}
		if err := json.Unmarshal([]byte(s), &obj); err != nil {
			return nil, errors.Errorf("map value is badly formatted: %s", s)
		}
		return obj, nil
	default:
		return s, nil
	}
}
