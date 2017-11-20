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
	"errors"
	"fmt"

	"github.com/google/go-jsonnet/ast"
)

func visitComponentsObj(component, snippet string) (*ast.Object, error) {
	root, err := astRoot(component, snippet)
	if err != nil {
		return nil, err
	}

	var componentsObj *ast.Object
	err = visit(root, &componentsObj)
	if err != nil {
		return nil, err
	}
	if componentsObj == nil {
		return nil, fmt.Errorf("Could not find object node: %s", componentsID)
	}

	return componentsObj, nil
}

func visit(node ast.Node, componentsObj **ast.Object) error {
	switch n := node.(type) {
	case *ast.Object:
		for _, field := range n.Fields {
			if field.Id != nil && *field.Id == componentsID {
				c, isObj := field.Expr2.(*ast.Object)
				if !isObj {
					return fmt.Errorf("Expected %s node type to be object", componentsID)
				}
				*componentsObj = c
				return nil
			}
			err := visitObjectField(field, componentsObj)
			if err != nil {
				return err
			}
		}
	case *ast.Apply:
		for _, arg := range n.Arguments.Positional {
			err := visit(arg, componentsObj)
			if err != nil {
				return err
			}
		}

		for _, arg := range n.Arguments.Named {
			err := visit(arg.Arg, componentsObj)
			if err != nil {
				return err
			}
		}
		return visit(n.Target, componentsObj)
	case *ast.ApplyBrace:
		err := visit(n.Left, componentsObj)
		if err != nil {
			return err
		}
		return visit(n.Right, componentsObj)
	case *ast.Array:
		for _, element := range n.Elements {
			err := visit(element, componentsObj)
			if err != nil {
				return err
			}
		}
	case *ast.ArrayComp:
		err := visitCompSpec(n.Spec, componentsObj)
		if err != nil {
			return err
		}
		return visit(n.Body, componentsObj)
	case *ast.Assert:
		err := visit(n.Cond, componentsObj)
		if err != nil {
			return err
		}
		err = visit(n.Message, componentsObj)
		if err != nil {
			return err
		}
		return visit(n.Rest, componentsObj)
	case *ast.Binary:
		err := visit(n.Left, componentsObj)
		if err != nil {
			return err
		}
		return visit(n.Right, componentsObj)
	case *ast.Conditional:
		err := visit(n.BranchFalse, componentsObj)
		if err != nil {
			return err
		}
		err = visit(n.BranchTrue, componentsObj)
		if err != nil {
			return err
		}
		return visit(n.Cond, componentsObj)
	case *ast.Error:
		return visit(n.Expr, componentsObj)
	case *ast.Function:
		for _, p := range n.Parameters.Optional {
			err := visit(p.DefaultArg, componentsObj)
			if err != nil {
				return err
			}
		}
		return visit(n.Body, componentsObj)
	case *ast.Index:
		err := visit(n.Target, componentsObj)
		if err != nil {
			return err
		}
		return visit(n.Index, componentsObj)
	case *ast.Slice:
		err := visit(n.Target, componentsObj)
		if err != nil {
			return err
		}
		err = visit(n.BeginIndex, componentsObj)
		if err != nil {
			return err
		}
		err = visit(n.EndIndex, componentsObj)
		if err != nil {
			return err
		}
		return visit(n.Step, componentsObj)
	case *ast.Local:
		for _, bind := range n.Binds {
			err := visitLocalBind(bind, componentsObj)
			if err != nil {
				return err
			}
		}
		return visit(n.Body, componentsObj)
	case *ast.DesugaredObject:
		for _, assert := range n.Asserts {
			err := visit(assert, componentsObj)
			if err != nil {
				return err
			}
		}
		for _, field := range n.Fields {
			err := visitDesugaredObjectField(field, componentsObj)
			if err != nil {
				return err
			}
		}
	case *ast.ObjectComp:
		for _, field := range n.Fields {
			err := visitObjectField(field, componentsObj)
			if err != nil {
				return err
			}
		}
		err := visitCompSpec(n.Spec, componentsObj)
		if err != nil {
			return err
		}
	case *ast.SuperIndex:
		return visit(n.Index, componentsObj)
	case *ast.InSuper:
		return visit(n.Index, componentsObj)
	case *ast.Unary:
		return visit(n.Expr, componentsObj)
	case *ast.Import:
	case *ast.ImportStr:
	case *ast.Dollar:
	case *ast.LiteralBoolean:
	case *ast.LiteralNull:
	case *ast.LiteralNumber:
	case *ast.LiteralString:
	case *ast.Self:
	case *ast.Var:
	case nil:
		return nil
	default:
		return errors.New("Unsupported ast.Node type found")
	}

	return nil
}

func visitCompSpec(node ast.ForSpec, componentsObj **ast.Object) error {
	if node.Outer != nil {
		err := visitCompSpec(*node.Outer, componentsObj)
		if err != nil {
			return err
		}
	}

	for _, ifspec := range node.Conditions {
		err := visit(ifspec.Expr, componentsObj)
		if err != nil {
			return err
		}
	}
	return visit(node.Expr, componentsObj)
}

func visitObjectField(node ast.ObjectField, componentsObj **ast.Object) error {
	if node.Method != nil {
		err := visit(node.Method, componentsObj)
		if err != nil {
			return err
		}
	}

	if node.Params != nil {
		for _, p := range node.Params.Optional {
			err := visit(p.DefaultArg, componentsObj)
			if err != nil {
				return err
			}
		}
	}

	err := visit(node.Expr1, componentsObj)
	if err != nil {
		return err
	}
	err = visit(node.Expr2, componentsObj)
	if err != nil {
		return err
	}
	return visit(node.Expr3, componentsObj)
}

func visitDesugaredObjectField(node ast.DesugaredObjectField, componentsObj **ast.Object) error {
	err := visit(node.Name, componentsObj)
	if err != nil {
		return err
	}
	return visit(node.Body, componentsObj)
}

func visitLocalBind(node ast.LocalBind, componentsObj **ast.Object) error {
	if node.Fun != nil {
		err := visit(node.Fun, componentsObj)
		if err != nil {
			return err
		}
	}
	return visit(node.Body, componentsObj)
}
