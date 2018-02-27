package astext

import "github.com/google/go-jsonnet/ast"

// ObjectFields is a slice of ObjectField.
type ObjectFields []ObjectField

// ObjectField wraps ast.ObjectField and adds commenting and the ability to
// be printed on one line.
type ObjectField struct {
	ast.ObjectField

	// Comment is a comment for the object field.
	Comment *Comment

	// Oneline prints this field on a single line.
	Oneline bool
}

// Object wraps ast.Object and adds the ability to be printed on one line.
type Object struct {
	ast.Object

	Fields []ObjectField

	// Oneline prints this field on a single line.
	Oneline bool
}
